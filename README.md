# CHCC - Change Claude Code API

A command-line tool for managing Claude Code API configurations with automatic environment variable management.

## Installation

### From Source

```bash
git clone https://github.com/Inasayang/chcc
cd chcc
go build -o chcc
```

## Quick Start

1. **Add your first Claude Code API configuration:**
   ```bash
   chcc add -n "Official" -u "https://api.anthropic.com" -t "sk-your-anthropic-api-key"
   ```

2. **Add a proxy or alternative endpoint:**
   ```bash
   chcc add -n "Proxy" -u "https://your-proxy.example.com" -t "your-proxy-token"
   ```

3. **List all Claude Code API configurations:**
   ```bash
   chcc list
   # or simply
   chcc
   ```

4. **Switch to a different configuration:**
   ```bash
   chcc set-default -n "Proxy"
   ```

5. **Remove a configuration:**
   ```bash
   chcc remove -n "Old-Config"
   ```

## Dependencies

- [Cobra](https://github.com/spf13/cobra): CLI framework
- [YAML v2](https://gopkg.in/yaml.v2): YAML configuration parsing

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Changelog

### v1.0.0

- Initial release with add, list, set-default, and remove commands
- Cross-platform environment variable management
- User-level configuration storage
- Seamless Claude Code integration