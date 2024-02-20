package cmd

import (
	"github.com/spf13/cobra"

	"github.com/instill-ai/component/tools/compogen/pkg/gen"
)

func init() {
	genReadmeCmd := &cobra.Command{
		Use:  "readme [config dir] [target file]",
		Args: cobra.ExactArgs(2),

		Short: "Generate component README",
		Long: `Generates a README.mdx file that describes the purpose and usage of the component.

The first argument specifies the path to the component config directory, i.e., the directory holding the component definitions.
The second argument allows users to specify the path to the generated README file.`,

		RunE: wrapRun(func(cmd *cobra.Command, args []string) error {
			return gen.NewREADMEGenerator(args[0], args[1]).Generate()
		}),
	}

	rootCmd.AddCommand(genReadmeCmd)
}
