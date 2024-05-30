# OK - Check readme usage

compogen readme --help
cmp stdout want-help

compogen help readme
cmp stdout want-help

# NOK - Invalid positional args

! compogen readme
cmp stderr want-0-args

! compogen readme foo
cmp stderr want-1-arg

-- want-help --
Generates a README.mdx file that describes the purpose and usage of the component.

The first argument specifies the path to the component config directory, i.e., the directory holding the component definitions.
The second argument allows users to specify the path to the generated README file.

Usage:
  compogen readme [config dir] [target file] [flags]

Flags:
  -h, --help   help for readme
-- want-0-args --
Error: accepts 2 arg(s), received 0
Run 'compogen readme --help' for usage.
-- want-1-arg --
Error: accepts 2 arg(s), received 1
Run 'compogen readme --help' for usage.
