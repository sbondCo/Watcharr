<script lang="ts">
	import { pushState } from "$app/navigation";
	import { page } from "$app/state";
	import {
		entries,
		type HandbookEntryId,
		type HandbookEntry,
	} from "@/lib/handbook/entries";
	import Icon from "@/lib/Icon.svelte";

	let isOnHandbookRoute: boolean = $derived(page.route.id == "/(app)/handbook");
	let active: HandbookEntryId | undefined = $derived.by(() => {
		if (isOnHandbookRoute) {
			const p = page.url.searchParams.get("entry");
			if (!p) {
				console.debug("handbook: active: no search param.");
				return undefined;
			}
			console.debug("handbook: active:", p);
			return p as HandbookEntryId;
		}
		return page.state.handbook;
	});
	let activeEntry: HandbookEntry | undefined = $derived.by(() => {
		if (!active) {
			return;
		}
		for (let i = 0; i < entries.length; i++) {
			const g = entries[i];
			const f = g.entries.find((e) => e.id == active);
			if (f) {
				return f;
			}
		}
	});

	function setActive(a: HandbookEntryId | undefined) {
		active = a;

		const url = new URL(page.url);
		if (active) {
			url.searchParams.set("entry", active);
		} else {
			url.searchParams.delete("entry");
		}
		console.debug("handbook: setActive:", url);
		// TODO: sveltekit v3 migration: Current pushState is broken, so when we can move to using
		// the new `goto` func with `shallow`, we can rely on page.shallow.url or whatever its called
		// to get the actual browser url, which means we can rely on the `entry` search param for realsies
		// (back/forward buttons in browser will then work in here and make the change).
		// eslint-disable-next-line svelte/no-navigation-without-resolve
		pushState(url, page.state);
	}
</script>

<svelte:head>
	<title>{activeEntry ? `${activeEntry.title} - ` : ""}Handbook</title>
</svelte:head>

<div class="content">
	<div class="inner">
		<div class="entry" class:is-on-handbook-route={isOnHandbookRoute}>
			{#if activeEntry}
				<div class="top-bar">
					<button class="back" onclick={() => setActive(undefined)}>
						<Icon i="chevron" facing="left" wh={18} />
						<span>Back</span>
					</button>

					<span class="title">{activeEntry.title}</span>
				</div>

				{#await activeEntry.comp() then module}
					{@const EntryComponent = module.default}
					<EntryComponent />
				{:catch error}
					<p>Couldn't load component: {error.message}</p>
				{/await}
			{:else}
				<h2 class="norm">
					<!-- NOTE for later self: This page is in beta since its navigation is slightly broken.
					 after sveltekit v3 migration we should be able to fix the shallow routing with the new goto func. -->
					Welcome to your handbook. <span style="font-size: 12px">Beta</span>
				</h2>

				<p>
					Inside, you will find useful information in regard to using Watcharr.
				</p>
				<p>
					Part of Watcharr's mission is to be extra user-friendly (aka
					intuitive), so hopefully you don't need to reference guides much, but
					they may come in handy, which is why this handbook was made.
				</p>

				<div class="entries">
					{#each entries as group (group.title)}
						<h2 class="plain">{group.title}</h2>
						<div class="btns">
							{#each group.entries as ent (ent.id)}
								<button class:plain={true} onclick={() => setActive(ent.id)}>
									{ent.title}
								</button>
							{/each}
						</div>
					{/each}
				</div>
			{/if}
		</div>
	</div>
</div>

<style lang="scss">
	$pad: 40px;

	.content {
		display: flex;
		width: 100%;
		justify-content: center;

		.inner {
			max-width: 1800px;
			display: flex;
			flex-flow: column;
			gap: 20px;
		}
	}

	.entry {
		width: 100dvw;
		padding-right: $pad;
		padding-left: $pad;

		/* When embedded in a Modal, different width/padding needed. */
		&:not(.is-on-handbook-route) {
			width: 100%;
			padding-right: 0;
			padding-left: 0;
		}

		div.top-bar {
			display: flex;
			flex-flow: row;
			align-items: center;
			gap: 15px;
			margin-bottom: 20px;
			padding: 10px 0;

			button.back {
				display: none;
				align-self: start;
				width: max-content;
				gap: 5px;
			}

			.title {
				font-size: 23px;
				font-weight: bold;
			}
		}

		/* Only sticky on handbook route. */
		&.is-on-handbook-route div.top-bar {
			position: sticky;
			top: 0px;
			transition: top 200ms ease-in-out;
			@include nav-blur;

			button.back {
				display: flex;
			}
		}

		.entries {
			display: flex;
			flex-flow: column;

			.btns {
				display: flex;
				flex-flow: row;
				gap: 10px;
				margin-bottom: 20px;

				button {
					padding: 20px;
					background-color: $accent-color;
					border-radius: 10px;
					font-size: 16px;
					transition:
						background-color 150ms ease,
						color 150ms ease;

					&:hover {
						color: $bg-color;
						background-color: $accent-color-hover;
					}
				}
			}
		}

		/* Styling for doc components */
		:global {
			p {
				margin-bottom: 15px;
			}

			ol,
			ul {
				padding-left: 32px;
			}

			img {
				width: auto;
				max-width: 100%;
			}

			code {
				background-color: $accent-color;
				border-radius: 3px;
				padding: 1px 3px;
				font-size: 14px;
			}

			a {
				color: blue;
			}

			h1,
			h2,
			h3,
			h4,
			h5 {
				font-family: unset;
			}

			h1,
			h2,
			h3 {
				margin-bottom: 20px;
			}
		}
	}

	:global(body.nav-shown) .entry.is-on-handbook-route div.top-bar {
		top: $nav-height;
	}
</style>
