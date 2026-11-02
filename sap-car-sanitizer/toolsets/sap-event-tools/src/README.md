# sap-event-tools

Scaffolded by `sam toolset init` (SAM CLI version 2.346.0).

## Authoring

To have your AI coding assistant (Claude Code, Cursor, ...) write the tool
with the correct `samtoolsdk` API, install the SAM authoring skills
once at your repo root:

```bash
sam ai-assistance skill install
```

Then edit `main.go`, replacing the scaffolded `greet` example
with your tool.

## Build

```bash
./build.sh
```

The Go SDK is vendored at `_sdk/samtoolsdk/` and wired in via
`go.mod`'s `replace` directive — no `go get` step is needed.
(The directory is named `_sdk` rather than the conventional
`vendor/` so Go does not switch to vendor-mode for the whole tool.)

Run `sam toolset sync` to refresh the vendored SDK after
upgrading the CLI.

## Layout

```
toolsets/sap-event-tools.yaml            # kind: toolset metadata header
toolsets/sap-event-tools/
  src/                             # this directory
    main.go                        # tool entry point (uses samtoolsdk)
    go.mod                         # replace samtoolsdk => ./_sdk/samtoolsdk
    _sdk/samtoolsdk/              # vendored SDK source (gitignored by default)
    build.sh / .bat                # cross-platform build scripts
    manifest.yaml                  # STR registration
  dist/                            # build output (gitignored)
  .sam-cache/                      # build cache (gitignored)
```
