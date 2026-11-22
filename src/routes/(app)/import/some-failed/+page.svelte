<!-- 
  /import/some-failed shows the user the
  failed imports so they can manually import
  them instead.

  TODO: It would be better if this page didn't exist
    and we instead just let users `re-try` on import/process
    table. If any failed, we can hide the success ones (and
    add a toggle to show successful ones), user can edit any
    failed rows and try import each again.
 -->

<script lang="ts">
	import { goto } from "$app/navigation";
	import { resolve } from "$app/paths";
	import Icon from "@/lib/Icon.svelte";
	import SpinnerTiny from "@/lib/SpinnerTiny.svelte";
	import { req } from "@/lib/util/api";
	import { notify } from "@/lib/util/notify";
	import { store } from "@/store.svelte";
	import { ImportResponseType, type ImportedList } from "@/types";
	import { onMount } from "svelte";

	let failed: ImportedList[] = $state([]);
	let successCount = $state(0);
	let ignoredCount = $state(0);
	// Names with an ignore request in flight, so only that rows button is
	// disabled rather than the whole list.
	let ignoring: string[] = $state([]);

	/**
	 * Give up on an entry for good. Nothing is imported, the decision is
	 * saved as an ignored match, and future imports skip the name instead of
	 * failing on it again.
	 */
	async function ignore(item: ImportedList) {
		const name = item.name;
		if (!name) {
			return;
		}
		ignoring = [...ignoring, name];
		try {
			await req.post("/import", { ...item, ignoreThisItem: true });
			failed = failed.filter((f) => f !== item);
			ignoredCount++;
			notify({ type: "success", text: `Ignoring ${name} from now on` });
		} catch (err) {
			console.error("some-failed: failed to ignore", err);
			notify({ type: "error", text: `Couldn't ignore ${name}` });
		}
		ignoring = ignoring.filter((n) => n !== name);
	}

	onMount(() => {
		if (store.parsedImportedList) {
			for (let i = 0; i < store.parsedImportedList.length; i++) {
				const item = store.parsedImportedList[i];
				$state.snapshot(item);
				if (
					item.state === ImportResponseType.IMPORT_FAILED ||
					item.state === ImportResponseType.IMPORT_NOTFOUND
				) {
					failed.push(item);
					failed = failed;
				} else if (item.state === ImportResponseType.IMPORT_IGNORED) {
					ignoredCount++;
				} else {
					successCount++;
				}
			}
			console.log("failedlen", failed.length);
		} else {
			goto(resolve("/import"));
		}
	});
</script>

<div class="content">
	<div class="inner">
		<h2>Some Content Failed To Import</h2>
		<h5 class="norm">
			You can search for the failed imports and manually add them.
		</h5>
		<h4 class="norm">
			{successCount} succeeded and {failed.length} failed{ignoredCount > 0
				? `, ${ignoredCount} ignored`
				: ""}.
		</h4>

		{#if failed}
			<ul>
				<!-- TODO: Fix this to use a keyed each somehow (need unique id for key) -->
				<!-- eslint-disable-next-line svelte/require-each-key -->
				{#each failed as l}
					<li>
						<span>{l.name}</span>
						<div class="ignore-wrap">
							{#if ignoring.includes(l.name ?? "")}
								<SpinnerTiny />
							{:else}
								<button
									title="Never try to import this name again"
									onclick={() => ignore(l)}
								>
									<Icon i="eye-closed" wh={18} />Ignore
								</button>
							{/if}
						</div>
					</li>
				{/each}
			</ul>
		{:else}
			No List
		{/if}
	</div>
</div>

<style lang="scss">
	.content {
		display: flex;
		width: 100%;
		justify-content: center;
		padding: 0 30px 30px 30px;

		.inner {
			display: flex;
			flex-flow: column;
			min-width: 400px;
			max-width: 600px;
			overflow: hidden;

			@media screen and (max-width: 420px) {
				min-width: 100%;
			}
		}
	}

	h4 {
		margin-top: 15px;
	}

	ul {
		display: flex;
		flex-flow: column;
		gap: 5px;
		margin: 10px;
		list-style: none;

		li {
			display: flex;
			flex-flow: row;
			align-items: center;
			padding: 10px;
			background-color: $accent-color;
			border-radius: 5px;

			a {
				margin-left: auto;

				button {
					width: max-content;
				}
			}

			.ignore-wrap {
				margin-left: auto;

				button {
					display: flex;
					align-items: center;
					gap: 3px;
					width: max-content;
				}
			}
		}
	}
</style>
