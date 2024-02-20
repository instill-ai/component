package cmd

import (
	"github.com/spf13/cobra"

	"github.com/instill-ai/component/tools/compogen/pkg/gen"
)

func init() {
	var isConnector, isOperator bool

	genReadmeCmd := &cobra.Command{
		Use:  "readme [config dir] [target file]",
		Args: cobra.ExactArgs(2),

		Short: "Generate component README",
		Long: `Generates a README.mdx file that describes the purpose and usage of the component.

The first argument specifies the path to the component config directory, i.e., the directory holding the component definitions.
The second argument allows users to specify the path to the generated README file.`,

		RunE: wrapRun(func(cmd *cobra.Command, args []string) error {
			var ct gen.ComponentType
			switch {
			case isConnector:
				ct = gen.ComponentTypeConnector
			case isOperator:
				ct = gen.ComponentTypeOperator
			}

			return gen.NewREADMEGenerator(args[0], args[1], ct).
				Generate()
		}),
	}

	genReadmeCmd.Flags().BoolVar(&isConnector, "connector", false, "Document connector component")
	genReadmeCmd.Flags().BoolVar(&isOperator, "operator", false, "Document operator component")
	genReadmeCmd.MarkFlagsOneRequired("connector", "operator")
	genReadmeCmd.MarkFlagsMutuallyExclusive("connector", "operator")

	rootCmd.AddCommand(genReadmeCmd)
}
