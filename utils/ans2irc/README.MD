# ANSI Art to IRC Format Converter

This Go program is designed to convert ANSI art files into a format suitable for IRC chat clients. It has been tested on *.ANS files from [16colo.rs](https://16colo.rs/), with modern terminal art considered out of scope for this tool.

## Installation and Usage
1. Ensure you have the latest Go installed on your system.
2. Clone or download this repository.
3. Run the following command to compile the program:
   ```bash
   go build
   ```

    Or just run:
    ```bash
    go install github.com/ugjka/kline/utils/ans2irc@latest
    ```

4. Use the program by providing a filename of an ANSI art file as an argument:
   ```bash
   ./ans2irc example.ans > example.txt
   ```
   - The `-cols` flag allows you to specify the column count (80 being the default) in the ANSI artwork.

## Features
- Converts ANSI escape sequences to compatible IRC control characters.
- Replaces control characters with characters from the IBM437 character set.
- Handles cursor positioning, text formatting, and color changes to translate ANSI art accurately to IRC.

## Notes
- In case of any errors or incomplete ANSI sequences, the program will provide feedback on the standard output.
- Some ANSI commands or parameters may not be fully supported or handled. Any encountered issues will be displayed in the standard output.

Feel free to use this tool to convert ANSI art files into IRC-friendly format for sharing and displaying in chat environments.
