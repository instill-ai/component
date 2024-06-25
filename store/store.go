package store

import (
	"fmt"
	"sync"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/ai/archetypeai/v0"
	"github.com/instill-ai/component/ai/huggingface/v0"
	"github.com/instill-ai/component/ai/instill/v0"
	"github.com/instill-ai/component/ai/openai/v0"
	"github.com/instill-ai/component/ai/stabilityai/v0"
	"github.com/instill-ai/component/application/googlesearch/v0"
	"github.com/instill-ai/component/application/mail/v0"
	"github.com/instill-ai/component/application/numbers/v0"
	"github.com/instill-ai/component/application/restapi/v0"
	"github.com/instill-ai/component/application/slack/v0"
	"github.com/instill-ai/component/application/website/v0"
	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/data/bigquery/v0"
	"github.com/instill-ai/component/data/googlecloudstorage/v0"
	"github.com/instill-ai/component/data/pinecone/v0"
	"github.com/instill-ai/component/data/redis/v0"
	"github.com/instill-ai/component/operator/base64/v0"
	"github.com/instill-ai/component/operator/document/v0"
	"github.com/instill-ai/component/operator/image/v0"
	"github.com/instill-ai/component/operator/json/v0"
	"github.com/instill-ai/component/operator/text/v0"

	pb "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
)

var (
	once      sync.Once
	compStore *Store
)

// Store holds in-memory information about the initialized components.
type Store struct {
	componentUIDs   []uuid.UUID
	componentUIDMap map[uuid.UUID]*component
	componentIDMap  map[string]*component
}

type component struct {
	comp base.IComponent
}

// ComponentSecrets contains the global config secrets of each
// implemented component (referenced by ID). Components may use these secrets
// to skip the component configuration step and have a ready-to-run
// config.
type ComponentSecrets map[string]map[string]any

// Init initializes the components implemented in this repository and loads
// their information to memory.
func Init(
	logger *zap.Logger,
	secrets ComponentSecrets,
	usageHandlerCreators map[string]base.UsageHandlerCreator,
) *Store {
	baseComp := base.Component{Logger: logger}

	once.Do(func() {
		compStore = &Store{
			componentUIDMap: map[uuid.UUID]*component{},
			componentIDMap:  map[string]*component{},
		}
		compStore.Import(base64.Init(baseComp))
		compStore.Import(json.Init(baseComp))
		compStore.Import(image.Init(baseComp))
		compStore.Import(text.Init(baseComp))
		compStore.Import(document.Init(baseComp))

		{
			// StabilityAI
			conn := stabilityai.Init(baseComp)

			// Secret doesn't allow hyphens
			conn = conn.WithSecrets(secrets["stabilityai"]).
				WithUsageHandlerCreator(usageHandlerCreators[conn.GetID()])
			compStore.Import(conn)
		}

		compStore.Import(instill.Init(baseComp))
		compStore.Import(huggingface.Init(baseComp))

		{
			// OpenAI
			conn := openai.Init(baseComp)
			conn = conn.WithSecrets(secrets[conn.GetID()]).
				WithUsageHandlerCreator(usageHandlerCreators[conn.GetID()])
			compStore.Import(conn)
		}

		compStore.Import(archetypeai.Init(baseComp))
		compStore.Import(numbers.Init(baseComp))
		compStore.Import(bigquery.Init(baseComp))
		compStore.Import(googlecloudstorage.Init(baseComp))
		compStore.Import(googlesearch.Init(baseComp))
		compStore.Import(pinecone.Init(baseComp))
		compStore.Import(redis.Init(baseComp))
		compStore.Import(restapi.Init(baseComp))
		compStore.Import(website.Init(baseComp))
		compStore.Import(slack.Init(baseComp))
		compStore.Import(mail.Init(baseComp))

	})
	return compStore
}

// Import loads the component definitions into memory.
func (s *Store) Import(comp base.IComponent) {
	c := &component{comp: comp}
	s.componentUIDMap[comp.GetUID()] = c
	s.componentIDMap[comp.GetID()] = c
	s.componentUIDs = append(s.componentUIDs, comp.GetUID())
}

// CreateExecution initializes the execution of a component given its UID.
func (s *Store) CreateExecution(defUID uuid.UUID, sysVars map[string]any, setup *structpb.Struct, task string) (*base.ExecutionWrapper, error) {
	if c, ok := s.componentUIDMap[defUID]; ok {
		return c.comp.CreateExecution(sysVars, setup, task)
	}
	return nil, fmt.Errorf("component definition not found")
}

// GetDefinitionByUID returns a component definition by its UID.
func (s *Store) GetDefinitionByUID(defUID uuid.UUID, sysVars map[string]any, compConfig *base.ComponentConfig) (*pb.ComponentDefinition, error) {
	if c, ok := s.componentUIDMap[defUID]; ok {
		def, err := c.comp.GetDefinition(sysVars, compConfig)
		if err != nil {
			return nil, err
		}
		return proto.Clone(def).(*pb.ComponentDefinition), err
	}
	return nil, fmt.Errorf("component definition not found")
}

// GetDefinitionByID returns a component definition by its ID.
func (s *Store) GetDefinitionByID(defID string, sysVars map[string]any, compConfig *base.ComponentConfig) (*pb.ComponentDefinition, error) {
	if c, ok := s.componentIDMap[defID]; ok {
		def, err := c.comp.GetDefinition(sysVars, compConfig)
		if err != nil {
			return nil, err
		}
		return proto.Clone(def).(*pb.ComponentDefinition), err
	}
	return nil, fmt.Errorf("component definition not found")
}

// ListDefinitions returns all the loaded component definitions.
func (s *Store) ListDefinitions(sysVars map[string]any, returnTombstone bool) []*pb.ComponentDefinition {
	defs := []*pb.ComponentDefinition{}
	for _, uid := range s.componentUIDs {
		c := s.componentUIDMap[uid]
		def, err := c.comp.GetDefinition(sysVars, nil)
		if err == nil {
			if !def.Tombstone || returnTombstone {
				defs = append(defs, proto.Clone(def).(*pb.ComponentDefinition))
			}
		}
	}
	return defs
}

func (s *Store) IsSecretField(defUID uuid.UUID, target string) (bool, error) {
	if c, ok := s.componentUIDMap[defUID]; ok {
		return c.comp.IsSecretField(target), nil
	}
	return false, fmt.Errorf("component definition not found")
}
