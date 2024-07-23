# Logger

## Description

This tool is used to log messages to the console for a better debugging experience.

## Features
 - Automatically expand objects and arrays to make it easier to read.
 - Provide a clear interface to log messages in colors.
 - While no name is provided, the tool will utilize the file name, the function name and the line of file as logger information.

## Usage

1. Import the package
   ```golang
    import 	"github.com/instill-ai/component/tools/logger"
   ```
2. Create a new logger session
   ```golang
   logger := logger.SessionStart("Logger Name", VerboseLevel)
   defer logger.SessionEnd()
   ```
    - The first parameter is the name of the logger. If an empty name is provided, the logger will use the file name and the function name as session id.
    - The second parameter is the verbose level of the logger. The logger will only log messages with the level equal to or lower than the verbose level. The verbose level is an integer, and the logger provides the following levels:
        - Static : No message will be logged
        - Error  : Only error messages will be logged
        - Warn   : Error and warning messages will be logged
        - Develop: All messages will be logged
3. Use the logger to log messages
    - Single message
    ```golang
    logger.Info("This is an info message")
    logger.Warn("This is a warning message")
    logger.Success("This is a success message")
    logger.Error("This is an error message")
    ```
    - Messages with a title (if the first parameter is type of string)
    ```golang
    logger.Info("Title 1", "This is an info message")
    logger.Info("Title 2", structA, structB, structC)
    // Output:
    // - Title 1: This is an info message
    // - Title 2: [
    //   structA {
    //     ...
    //   },
    //   structB {
    //     ...
    //   },
    //   ...
    // ]
    ...
    ```
    - Message with only non-string data
    ```golang
    logger.Info(structA, structB, structC)
    // Output:
    // - functionName:lineNumber : [
    //   structA {
    //     ...
    //   },
    //   structB {
    //     ...
    //   },
    //   ...
    // ]
    ...
    ```

## Note
- The default expand level is 5. If you want to expand more or less, you can change the value by calling the function `logger.SetMaxDepth(level int)`.
- The default indent symbol is "  " (2 spaces) . You can change the indent symbol by calling the function `logger.SetIndentSymbol(symbol string)`.
