declare module 'astro:content' {
	interface Render {
		'.mdx': Promise<{
			Content: import('astro').MarkdownInstance<{}>['Content'];
			headings: import('astro').MarkdownHeading[];
			remarkPluginFrontmatter: Record<string, any>;
		}>;
	}
}

declare module 'astro:content' {
	interface RenderResult {
		Content: import('astro/runtime/server/index.js').AstroComponentFactory;
		headings: import('astro').MarkdownHeading[];
		remarkPluginFrontmatter: Record<string, any>;
	}
	interface Render {
		'.md': Promise<RenderResult>;
	}

	export interface RenderedContent {
		html: string;
		metadata?: {
			imagePaths: Array<string>;
			[key: string]: unknown;
		};
	}
}

declare module 'astro:content' {
	type Flatten<T> = T extends { [K: string]: infer U } ? U : never;

	export type CollectionKey = keyof AnyEntryMap;
	export type CollectionEntry<C extends CollectionKey> = Flatten<AnyEntryMap[C]>;

	export type ContentCollectionKey = keyof ContentEntryMap;
	export type DataCollectionKey = keyof DataEntryMap;

	type AllValuesOf<T> = T extends any ? T[keyof T] : never;
	type ValidContentEntrySlug<C extends keyof ContentEntryMap> = AllValuesOf<
		ContentEntryMap[C]
	>['slug'];

	/** @deprecated Use `getEntry` instead. */
	export function getEntryBySlug<
		C extends keyof ContentEntryMap,
		E extends ValidContentEntrySlug<C> | (string & {}),
	>(
		collection: C,
		// Note that this has to accept a regular string too, for SSR
		entrySlug: E,
	): E extends ValidContentEntrySlug<C>
		? Promise<CollectionEntry<C>>
		: Promise<CollectionEntry<C> | undefined>;

	/** @deprecated Use `getEntry` instead. */
	export function getDataEntryById<C extends keyof DataEntryMap, E extends keyof DataEntryMap[C]>(
		collection: C,
		entryId: E,
	): Promise<CollectionEntry<C>>;

	export function getCollection<C extends keyof AnyEntryMap, E extends CollectionEntry<C>>(
		collection: C,
		filter?: (entry: CollectionEntry<C>) => entry is E,
	): Promise<E[]>;
	export function getCollection<C extends keyof AnyEntryMap>(
		collection: C,
		filter?: (entry: CollectionEntry<C>) => unknown,
	): Promise<CollectionEntry<C>[]>;

	export function getEntry<
		C extends keyof ContentEntryMap,
		E extends ValidContentEntrySlug<C> | (string & {}),
	>(entry: {
		collection: C;
		slug: E;
	}): E extends ValidContentEntrySlug<C>
		? Promise<CollectionEntry<C>>
		: Promise<CollectionEntry<C> | undefined>;
	export function getEntry<
		C extends keyof DataEntryMap,
		E extends keyof DataEntryMap[C] | (string & {}),
	>(entry: {
		collection: C;
		id: E;
	}): E extends keyof DataEntryMap[C]
		? Promise<DataEntryMap[C][E]>
		: Promise<CollectionEntry<C> | undefined>;
	export function getEntry<
		C extends keyof ContentEntryMap,
		E extends ValidContentEntrySlug<C> | (string & {}),
	>(
		collection: C,
		slug: E,
	): E extends ValidContentEntrySlug<C>
		? Promise<CollectionEntry<C>>
		: Promise<CollectionEntry<C> | undefined>;
	export function getEntry<
		C extends keyof DataEntryMap,
		E extends keyof DataEntryMap[C] | (string & {}),
	>(
		collection: C,
		id: E,
	): E extends keyof DataEntryMap[C]
		? Promise<DataEntryMap[C][E]>
		: Promise<CollectionEntry<C> | undefined>;

	/** Resolve an array of entry references from the same collection */
	export function getEntries<C extends keyof ContentEntryMap>(
		entries: {
			collection: C;
			slug: ValidContentEntrySlug<C>;
		}[],
	): Promise<CollectionEntry<C>[]>;
	export function getEntries<C extends keyof DataEntryMap>(
		entries: {
			collection: C;
			id: keyof DataEntryMap[C];
		}[],
	): Promise<CollectionEntry<C>[]>;

	export function render<C extends keyof AnyEntryMap>(
		entry: AnyEntryMap[C][string],
	): Promise<RenderResult>;

	export function reference<C extends keyof AnyEntryMap>(
		collection: C,
	): import('astro/zod').ZodEffects<
		import('astro/zod').ZodString,
		C extends keyof ContentEntryMap
			? {
					collection: C;
					slug: ValidContentEntrySlug<C>;
				}
			: {
					collection: C;
					id: keyof DataEntryMap[C];
				}
	>;
	// Allow generic `string` to avoid excessive type errors in the config
	// if `dev` is not running to update as you edit.
	// Invalid collection names will be caught at build time.
	export function reference<C extends string>(
		collection: C,
	): import('astro/zod').ZodEffects<import('astro/zod').ZodString, never>;

	type ReturnTypeOrOriginal<T> = T extends (...args: any[]) => infer R ? R : T;
	type InferEntrySchema<C extends keyof AnyEntryMap> = import('astro/zod').infer<
		ReturnTypeOrOriginal<Required<ContentConfig['collections'][C]>['schema']>
	>;

	type ContentEntryMap = {
		"docs": {
"404.md": {
	id: "404.md";
  slug: "404";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".md"] };
"api-keys.mdx": {
	id: "api-keys.mdx";
  slug: "api-keys";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"billing.mdx": {
	id: "billing.mdx";
  slug: "billing";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"configuration.mdx": {
	id: "configuration.mdx";
  slug: "configuration";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"declarative-builder.mdx": {
	id: "declarative-builder.mdx";
  slug: "declarative-builder";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"file-system-operations.mdx": {
	id: "file-system-operations.mdx";
  slug: "file-system-operations";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"getting-started.mdx": {
	id: "getting-started.mdx";
  slug: "getting-started";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"git-operations.mdx": {
	id: "git-operations.mdx";
  slug: "git-operations";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"index.mdx": {
	id: "index.mdx";
  slug: "index";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"language-server-protocol.mdx": {
	id: "language-server-protocol.mdx";
  slug: "language-server-protocol";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"limits.mdx": {
	id: "limits.mdx";
  slug: "limits";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"linked-accounts.mdx": {
	id: "linked-accounts.mdx";
  slug: "linked-accounts";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"log-streaming.mdx": {
	id: "log-streaming.mdx";
  slug: "log-streaming";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"mcp.mdx": {
	id: "mcp.mdx";
  slug: "mcp";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"organizations.mdx": {
	id: "organizations.mdx";
  slug: "organizations";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"preview-and-authentication.mdx": {
	id: "preview-and-authentication.mdx";
  slug: "preview-and-authentication";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"process-code-execution.mdx": {
	id: "process-code-execution.mdx";
  slug: "process-code-execution";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"python-sdk/async/async-daytona.mdx": {
	id: "python-sdk/async/async-daytona.mdx";
  slug: "python-sdk/async/async-daytona";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"python-sdk/async/async-file-system.mdx": {
	id: "python-sdk/async/async-file-system.mdx";
  slug: "python-sdk/async/async-file-system";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"python-sdk/async/async-git.mdx": {
	id: "python-sdk/async/async-git.mdx";
  slug: "python-sdk/async/async-git";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"python-sdk/async/async-lsp-server.mdx": {
	id: "python-sdk/async/async-lsp-server.mdx";
  slug: "python-sdk/async/async-lsp-server";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"python-sdk/async/async-object-storage.mdx": {
	id: "python-sdk/async/async-object-storage.mdx";
  slug: "python-sdk/async/async-object-storage";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"python-sdk/async/async-sandbox.mdx": {
	id: "python-sdk/async/async-sandbox.mdx";
  slug: "python-sdk/async/async-sandbox";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"python-sdk/async/async-volume.mdx": {
	id: "python-sdk/async/async-volume.mdx";
  slug: "python-sdk/async/async-volume";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"python-sdk/common/charts.mdx": {
	id: "python-sdk/common/charts.mdx";
  slug: "python-sdk/common/charts";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"python-sdk/common/errors.mdx": {
	id: "python-sdk/common/errors.mdx";
  slug: "python-sdk/common/errors";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"python-sdk/common/image.mdx": {
	id: "python-sdk/common/image.mdx";
  slug: "python-sdk/common/image";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"python-sdk/index.mdx": {
	id: "python-sdk/index.mdx";
  slug: "python-sdk";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"python-sdk/sync/daytona.mdx": {
	id: "python-sdk/sync/daytona.mdx";
  slug: "python-sdk/sync/daytona";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"python-sdk/sync/file-system.mdx": {
	id: "python-sdk/sync/file-system.mdx";
  slug: "python-sdk/sync/file-system";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"python-sdk/sync/git.mdx": {
	id: "python-sdk/sync/git.mdx";
  slug: "python-sdk/sync/git";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"python-sdk/sync/lsp-server.mdx": {
	id: "python-sdk/sync/lsp-server.mdx";
  slug: "python-sdk/sync/lsp-server";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"python-sdk/sync/object-storage.mdx": {
	id: "python-sdk/sync/object-storage.mdx";
  slug: "python-sdk/sync/object-storage";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"python-sdk/sync/process.mdx": {
	id: "python-sdk/sync/process.mdx";
  slug: "python-sdk/sync/process";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"python-sdk/sync/sandbox.mdx": {
	id: "python-sdk/sync/sandbox.mdx";
  slug: "python-sdk/sync/sandbox";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"python-sdk/sync/volume.mdx": {
	id: "python-sdk/sync/volume.mdx";
  slug: "python-sdk/sync/volume";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"regions.mdx": {
	id: "regions.mdx";
  slug: "regions";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"sandbox-management.mdx": {
	id: "sandbox-management.mdx";
  slug: "sandbox-management";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"snapshots.mdx": {
	id: "snapshots.mdx";
  slug: "snapshots";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"tools/api.mdx": {
	id: "tools/api.mdx";
  slug: "tools/api";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"tools/cli.mdx": {
	id: "tools/cli.mdx";
  slug: "tools/cli";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"typescript-sdk/charts.mdx": {
	id: "typescript-sdk/charts.mdx";
  slug: "typescript-sdk/charts";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"typescript-sdk/daytona.mdx": {
	id: "typescript-sdk/daytona.mdx";
  slug: "typescript-sdk/daytona";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"typescript-sdk/errors.mdx": {
	id: "typescript-sdk/errors.mdx";
  slug: "typescript-sdk/errors";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"typescript-sdk/execute-response.mdx": {
	id: "typescript-sdk/execute-response.mdx";
  slug: "typescript-sdk/execute-response";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"typescript-sdk/file-system.mdx": {
	id: "typescript-sdk/file-system.mdx";
  slug: "typescript-sdk/file-system";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"typescript-sdk/git.mdx": {
	id: "typescript-sdk/git.mdx";
  slug: "typescript-sdk/git";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"typescript-sdk/image.mdx": {
	id: "typescript-sdk/image.mdx";
  slug: "typescript-sdk/image";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"typescript-sdk/index.mdx": {
	id: "typescript-sdk/index.mdx";
  slug: "typescript-sdk";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"typescript-sdk/lsp-server.mdx": {
	id: "typescript-sdk/lsp-server.mdx";
  slug: "typescript-sdk/lsp-server";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"typescript-sdk/object-storage.mdx": {
	id: "typescript-sdk/object-storage.mdx";
  slug: "typescript-sdk/object-storage";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"typescript-sdk/process.mdx": {
	id: "typescript-sdk/process.mdx";
  slug: "typescript-sdk/process";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"typescript-sdk/sandbox.mdx": {
	id: "typescript-sdk/sandbox.mdx";
  slug: "typescript-sdk/sandbox";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"typescript-sdk/snapshot.mdx": {
	id: "typescript-sdk/snapshot.mdx";
  slug: "typescript-sdk/snapshot";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"typescript-sdk/volume.mdx": {
	id: "typescript-sdk/volume.mdx";
  slug: "typescript-sdk/volume";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"volumes.mdx": {
	id: "volumes.mdx";
  slug: "volumes";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
"web-terminal.mdx": {
	id: "web-terminal.mdx";
  slug: "web-terminal";
  body: string;
  collection: "docs";
  data: InferEntrySchema<"docs">
} & { render(): Render[".mdx"] };
};
"legacy-docs": {
"about/architecture.mdx": {
	id: "about/architecture.mdx";
  slug: "about/architecture";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"about/getting-started.mdx": {
	id: "about/getting-started.mdx";
  slug: "about/getting-started";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"about/what-is-daytona.mdx": {
	id: "about/what-is-daytona.mdx";
  slug: "about/what-is-daytona";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"configuration/api-keys.mdx": {
	id: "configuration/api-keys.mdx";
  slug: "configuration/api-keys";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"configuration/container-registries.mdx": {
	id: "configuration/container-registries.mdx";
  slug: "configuration/container-registries";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"configuration/git-providers.mdx": {
	id: "configuration/git-providers.mdx";
  slug: "configuration/git-providers";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"configuration/providers.mdx": {
	id: "configuration/providers.mdx";
  slug: "configuration/providers";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"configuration/server.mdx": {
	id: "configuration/server.mdx";
  slug: "configuration/server";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"configuration/target-config.mdx": {
	id: "configuration/target-config.mdx";
  slug: "configuration/target-config";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"configuration/workspace-templates.mdx": {
	id: "configuration/workspace-templates.mdx";
  slug: "configuration/workspace-templates";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"index.mdx": {
	id: "index.mdx";
  slug: "index";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"installation/installation.mdx": {
	id: "installation/installation.mdx";
  slug: "installation/installation";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"installation/method/homebrew.mdx": {
	id: "installation/method/homebrew.mdx";
  slug: "installation/method/homebrew";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"installation/method/nix.mdx": {
	id: "installation/method/nix.mdx";
  slug: "installation/method/nix";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"installation/method/script-powershell.mdx": {
	id: "installation/method/script-powershell.mdx";
  slug: "installation/method/script-powershell";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"installation/method/script-unix.mdx": {
	id: "installation/method/script-unix.mdx";
  slug: "installation/method/script-unix";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"installation/method/uninstall-windows.mdx": {
	id: "installation/method/uninstall-windows.mdx";
  slug: "installation/method/uninstall-windows";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"installation/method/uninstall.mdx": {
	id: "installation/method/uninstall.mdx";
  slug: "installation/method/uninstall";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"misc/telemetry.mdx": {
	id: "misc/telemetry.mdx";
  slug: "misc/telemetry";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"misc/troubleshooting.mdx": {
	id: "misc/troubleshooting.mdx";
  slug: "misc/troubleshooting";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"sdk/file-system-operations.mdx": {
	id: "sdk/file-system-operations.mdx";
  slug: "sdk/file-system-operations";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"sdk/git-operations.mdx": {
	id: "sdk/git-operations.mdx";
  slug: "sdk/git-operations";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"sdk/language-server-protocol.mdx": {
	id: "sdk/language-server-protocol.mdx";
  slug: "sdk/language-server-protocol";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"sdk/process-code-execution.mdx": {
	id: "sdk/process-code-execution.mdx";
  slug: "sdk/process-code-execution";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"sdk/sandbox-management.mdx": {
	id: "sdk/sandbox-management.mdx";
  slug: "sdk/sandbox-management";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"tools/api.mdx": {
	id: "tools/api.mdx";
  slug: "tools/api";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"tools/cli.mdx": {
	id: "tools/cli.mdx";
  slug: "tools/cli";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"tools/docker-extension.mdx": {
	id: "tools/docker-extension.mdx";
  slug: "tools/docker-extension";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"usage/agent-toolbox.mdx": {
	id: "usage/agent-toolbox.mdx";
  slug: "usage/agent-toolbox";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"usage/builders.mdx": {
	id: "usage/builders.mdx";
  slug: "usage/builders";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"usage/ide.mdx": {
	id: "usage/ide.mdx";
  slug: "usage/ide";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"usage/prebuilds.mdx": {
	id: "usage/prebuilds.mdx";
  slug: "usage/prebuilds";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"usage/runners.mdx": {
	id: "usage/runners.mdx";
  slug: "usage/runners";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"usage/samples.mdx": {
	id: "usage/samples.mdx";
  slug: "usage/samples";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"usage/targets.mdx": {
	id: "usage/targets.mdx";
  slug: "usage/targets";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
"usage/workspaces.mdx": {
	id: "usage/workspaces.mdx";
  slug: "usage/workspaces";
  body: string;
  collection: "legacy-docs";
  data: any
} & { render(): Render[".mdx"] };
};

	};

	type DataEntryMap = {
		"i18n": Record<string, {
  id: string;
  collection: "i18n";
  data: InferEntrySchema<"i18n">;
}>;

	};

	type AnyEntryMap = ContentEntryMap & DataEntryMap;

	export type ContentConfig = typeof import("../../src/content/config.js");
}
