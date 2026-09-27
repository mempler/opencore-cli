## OpenCore CLI v1.6.1

### Typegen reads the TypeScript AST

Typegen now walks each resource with the project's own TypeScript (5.x or 7.x) instead of matching
source text, so every listener the runtime registers is typed:

- Decorators are matched by the symbol they resolve to: named (`import { OnRPC }`), aliased
  (`import { OnRPC as Rpc }`), namespace and barrel re-exports of `@open-core/framework/server` and
  `/client` all count, not only `@Server.*` / `@Client.*`. A local decorator that happens to be
  called `OnNet` does not.
- Names come from the argument's literal type: `BankEvents.deposit` (`as const`), a local constant or a
  framework constant such as `SYSTEM_EVENTS.chat.message` (requires `@open-core/framework` with literal
  `SYSTEM_EVENTS` types). A warning is reported only when the name's type is plain `string`.
- The decorated method is taken from the AST: multi-line decorator arguments, stacked decorators and
  method names such as `do`, `delete` or `new` no longer skip a handler.
- `WebView.send()` payloads are typed from the checker whatever the expression (object literals,
  locals, class fields). Exported types are imported, others are expanded. A typed forwarder,
  `send<K extends keyof M>(name: K, data: M[K])`, registers one message per key of `M`, and its call
  sites are no longer read as sends of their own.
- Generated maps key constants by their value (`'bank:withdraw'`) instead of an `__Entry<typeof import(…)>`
  reference. Literal names produce the same output as before.

Typegen requires `typescript` in the project, like `build.typecheck`.

### Source validation

The mixed-decorator check only counts real decorators: a file whose comments, strings or type-only
imports mention `@Client.*` and `@Server.*` no longer cancels the build.

