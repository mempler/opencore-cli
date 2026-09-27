## OpenCore CLI v1.6.0

Requires `@open-core/framework` v1.2.0 or later for the generated types to take effect.

### Type generation for net events, RPCs, commands and WebView messages

`opencore build` and `opencore dev` now scan every resource and emit a `.opencore/opencore.gen.ts` that
registers its handlers with the framework's type registry. Emitting sites then autocomplete to the
declared names and check their payloads against the handler signatures:

| Source | Types |
|--------|-------|
| `@Server.OnNet` | `events.emit()` on the client (the `Player` parameter is dropped) |
| `@Client.OnNet` | `player.emit()` and `events.emit()` on the server |
| `@Server.OnRPC` / `@Client.OnRPC` | `rpc.call()` and `rpc.notify()` arguments, and the `call()` result |
| `@Server.Command` | the command map, with `description` and `usage` |
| `@Client.OnView` / `WebView.send()` | messages in both directions between the client and its WebViews |

Names can be string literals, file-local constants or imported constants such as `BankEvents.WITHDRAW`.
Every resource registers under its own key and the framework merges them, so a name declared in one
resource is known to all of them.

A resource with a WebView directory (`ui/`, `nui/` or the configured one) also gets a
`<view>/.opencore/opencore.gen.ts` exporting `UiCanSend` and `UiReceives` for the UI code. A `send()`
payload is typed when it forwards a parameter of the enclosing method; any other expression is typed
`unknown` and reported as a warning.

Generated files are rewritten only when their content changes, and `opencore dev` regenerates them on
every rebuild.

### Strict mode on by default

With `build.typegen.strict` (default `true`), an event, RPC or WebView message name that no handler
declares is a compile error. The framework's own `opencore:*` events are always accepted. Set
`strict: false` when a project emits names handled outside OpenCore, such as events of a Lua resource:
unknown names are then accepted while known ones keep their autocomplete and payload checks.

Existing projects get these errors in the editor and in `tsc` after upgrading. They only fail the
build when `build.typecheck` is enabled.

### `build.typecheck`

Bundling only strips types, so until now no type error could fail a build. With `build.typecheck: true`,
`opencore build` regenerates the types and runs the project's own `tsc --noEmit` before bundling, and
fails without writing any output. `opencore dev` reports the errors as warnings and keeps watching.

It is off by default so upgrading never breaks an existing build, and requires `typescript` installed
in the project.

### Configuration

```ts
build: {
  typecheck: true,
  typegen: {
    enabled: true, // default: true. When false, generated files are deleted and every signature is loose again
    strict: true, // default: true
  },
}
```

### Templates and tooling

- New projects enable `typecheck` and `typegen` in `opencore.config.ts`.
- The starter `tsconfig.json` includes the `.opencore` directories so the generated types are part of the program.
- `opencore create resource --with-nui` adds a `ui/tsconfig.json` that includes the view's generated types.
- `opencore doctor` reports the type generation mode, and warns when strict mode is on without `build.typecheck`.
