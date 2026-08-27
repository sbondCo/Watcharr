<!--
  /import/process is for processing the
  selected files data. Here it will be
  displayed and imported.
 -->

<script lang="ts">
	import { goto } from "$app/navigation";
	import Checkbox from "@/lib/Checkbox.svelte";
	import DropDown from "@/lib/DropDown.svelte";
	import Error from "@/lib/Error.svelte";
	import Icon from "@/lib/Icon.svelte";
	import Modal from "@/lib/Modal.svelte";
	import Poster from "@/lib/poster/Poster.svelte";
	import PosterList from "@/lib/poster/PosterList.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import SpinnerTiny from "@/lib/SpinnerTiny.svelte";
	import { sleep } from "@/lib/util/helpers";
	import { notify } from "@/lib/util/notify";
	import { store } from "@/store.svelte";
	import {
		ImportResponseType,
		getContentTypeFromMedia,
		type ImportResponse,
		type ImportedList,
		type Media,
		type WatchedStatus,
	} from "@/types";
	import { req } from "@/lib/util/api";
	import { onDestroy } from "svelte";
	import papa from "papaparse";
	import Status from "@/lib/Status.svelte";
	import { resolve } from "$app/paths";

	interface ImportedListItemMultiProblem {
		original: ImportedList;
		results: Media[];
		callback: (err: Error | string | undefined) => void;
	}

	let rList: ImportedList[] = $state([]);
	// When on, previously saved matches are ignored, so every ambiguous name
	// is asked about again. Deliberately not remembered between imports, it
	// is a choice for this run rather than a preference.
	let ignoreSavedMatches = $state(false);
	let isImporting = $state(false);
	let importText = $state("");
	let cancelled = $state(false);
	let importTableEl: HTMLTableElement | undefined = $state();

	onDestroy(() => {
		cancelled = true;
	});

	// This isn't reactive to avoid bugs (hopefully this isn't a bug)
	let dropDownSupportedTypes = (() => {
		let t = ["movie", "tv"];
		if (store.serverFeatures?.games) {
			t.push("game");
		}
		return t;
	})();

	// Set when current item being imported gets an IMPORT_MULTI
	// response, which then shows the modal for user to pick correct item.
	let importMultiItem: ImportedListItemMultiProblem | undefined = $state();
	// Set when user clicks 'Change Statuses' button for the change all
	// statuses modal.
	let changeAllStatusesModalCb:
		((newStatus?: WatchedStatus) => void) | undefined = $state();

	async function getList() {
		const list = store.importedList;
		if (!list) {
			console.log("import/process, no list, returning to /import");
			goto(resolve("/import"));
			return;
		}
		console.log("getList", list);
		if (list?.type === "text-list") {
			importText = "Text List";
			// Regex to match a year in between brackets,
			// which we assume is the release year of content.
			const yearRegex = new RegExp(/\([0-9]{4}\)/);
			// Matches supported rating between square brackets.
			// Supports (the sixes can be any number 0-9): [6], [6.6], and [10].
			const ratingRegex = new RegExp(/\[([0-9](?:\.[0-9])?|10)\]/);
			const s = list.data.split("\n");
			for (let i = 0; i < s.length; i++) {
				if (!s[i]) {
					console.warn("getList: ignoring entry:", s[i]);
					continue;
				}
				const l: ImportedList = {};
				l.name = s[i];
				// Try extracting a year.
				const year = l.name.match(yearRegex);
				if (year && year.length > 0) {
					l.year = Number(year[0].replaceAll(/\(|\)/g, ""));
					l.name = l.name.replace(yearRegex, "");
				}
				// Try extracting a rating.
				const rating = l.name.match(ratingRegex);
				if (rating && rating.length > 0) {
					// console.log("found rating", rating);
					l.rating = Number(rating[1]);
					if (typeof rating.index === "number") {
						l.name = l.name.slice(0, rating.index);
					}
				}
				l.name = l.name.trim();
				rList.push(l);
			}
		} else if (list?.type === "tmdb") {
			importText = "TMDB";
			const s = papa.parse(list.data.trim(), { header: true });
			console.debug("parsed csv", s);
			for (let i = 0; i < s.data.length; i++) {
				try {
					// eslint-disable-next-line @typescript-eslint/no-explicit-any
					const el = s.data[i] as any;
					if (el) {
						// Skip if no name or tmdb id
						if (!el.Name && !el["TMDb ID"]) {
							console.warn("Skipping item with no name or tmdb id", el);
							return;
						}
						const l: ImportedList = { name: el.Name };
						const year = el["Release Date"]
							? new Date(el["Release Date"])
							: undefined;
						if (year) {
							l.year = year.getFullYear();
						}
						if (el.Type === "movie" || el.Type === "tv") {
							l.type = el.Type;
						}
						if (el["TMDb ID"]) {
							l.tmdbId = Number(el["TMDb ID"]);
						}
						if (el["Your Rating"]) {
							l.rating = Math.floor(Number(el["Your Rating"]));
						}
						if (el["Date Rated"]) {
							l.ratingCustomDate = new Date(el["Date Rated"]);
						}
						rList.push(l);
					}
				} catch (err) {
					console.error("Failed to process an item!", err);
					notify({
						type: "error",
						text: "Failed to process an item!",
					});
				}
			}
		} else if (list?.type === "imdb") {
			// There are different types of imdb exports (ratings & watchlist/list),
			// there are common keys between these types that we use below so should
			// be okay with importing either.
			importText = "IMDb";
			// eslint-disable-next-line @typescript-eslint/no-explicit-any
			const s = papa.parse<any>(list.data.trim(), { header: true });
			console.debug("parsed csv", s);
			let anySkipped = false;
			// Sort so that episodes comes last, so they are imported last.
			s.data?.sort((a, b) => {
				const aType = a["Title Type"]?.toLowerCase();
				if (aType === "tv episode") {
					return 1;
				}
				const bType = b["Title Type"]?.toLowerCase();
				if (bType === "tv episode") {
					return -1;
				}
				return 0;
			});
			for (let i = 0; i < s.data.length; i++) {
				try {
					// eslint-disable-next-line @typescript-eslint/no-explicit-any
					const el = s.data[i] as any;
					if (el) {
						const imdbId = el["Const"];
						const type = el["Title Type"]?.toLowerCase();
						// Skip if no name or tmdb id
						if (!el.Title && !imdbId) {
							console.warn("Skipping item with no title or imdb id", el);
							anySkipped = true;
							continue;
						}
						if (!type) {
							console.warn("Skipping item with no type", el);
							anySkipped = true;
							continue;
						}
						const l: ImportedList = { name: el.Title };
						const year = el["Release Date"]
							? new Date(el["Release Date"])
							: undefined;
						if (year) {
							l.year = year.getFullYear();
						}

						switch (type) {
							case "movie":
							case "video":
							case "tv movie":
							case "short":
								l.type = "movie";
								break;
							case "tv series":
							case "tv mini series":
							case "tv special":
							case "tv short":
								l.type = "tv";
								break;
							case "tv episode":
								l.type = "tv_episode";
								break;
							default:
								console.warn(
									"Skipping item with invalid type",
									`(${type})`,
									el,
								);
								anySkipped = true;
								continue;
						}

						if (imdbId) {
							l.imdbId = imdbId;
						}
						if (el["Your Rating"]) {
							l.rating = Math.floor(Number(el["Your Rating"]));
						}
						if (el["Date Rated"]) {
							l.ratingCustomDate = new Date(el["Date Rated"]);
						}
						rList.push(l);
					}
				} catch (err) {
					console.error("Failed to process an item!", err);
					notify({
						type: "error",
						text: "Failed to process an item!",
					});
				}
			}
			notify({
				text: "Any tv episodes will be added to the bottom of the table, don't change the Type on these!",
				time: 20000,
			});
			if (anySkipped) {
				notify({
					type: "error",
					text: "Some items with invalid data may have been skipped (check source data for missing ids/titles or look in console for more details).",
				});
			}
		} else if (list?.type === "movary") {
			importText = "Movary";
			try {
				const s = JSON.parse(list.data);
				// Builds imported list in previous step for ease.
				rList = s;
			} catch (err) {
				console.error("Movary import processing failed!", err);
				notify({
					type: "error",
					text: "Processing failed!. Please report this issue if it persists.",
				});
			}
		} else if (list?.type === "watcharr") {
			importText = "Watcharr";
			try {
				const s = JSON.parse(list.data);
				// Builds imported list in previous step for ease.
				rList = s;
			} catch (err) {
				console.error("Watcharr import processing failed!", err);
				notify({
					type: "error",
					text: "Processing failed!. Please report this issue if it persists.",
				});
			}
		} else if (list?.type === "myanimelist") {
			importText = "MyAnimeList";
			try {
				const parser = new DOMParser();
				const doc = parser.parseFromString(list.data.trim(), "application/xml");
				const errorNode = doc.querySelector("parsererror");
				if (errorNode) {
					console.error("MyAnimeList parse error:", errorNode);
					notify({
						type: "error",
						text: "An error occurred while parsing your MyAnimeList export!",
					});
					return;
				}
				console.log(doc.documentElement.querySelectorAll("anime"));
				const animeNodes = doc.documentElement.querySelectorAll("anime");
				if (animeNodes?.length <= 0) {
					console.error("MyAnimeList: Found no anime nodes:", animeNodes);
					notify({
						type: "error",
						text: "We found no Anime entries in your export file!",
					});
					return;
				}
				for (let i = 0; i < animeNodes.length; i++) {
					const animeNode = animeNodes[i];
					const titleNode = animeNode.querySelector("series_title");
					console.debug("Processing anime:", titleNode?.textContent);
					if (!titleNode?.textContent) {
						console.error("No title found for an anime!", animeNode, titleNode);
						notify({
							type: "error",
							text: "An anime failed to import, a title was not found! Check console for more details.",
						});
						continue;
					}
					const l: ImportedList = { name: titleNode.textContent };
					const scoreNode = animeNode.querySelector("my_score");
					if (scoreNode?.textContent) {
						l.rating = Number(scoreNode.textContent);
					}
					const statusNode = animeNode.querySelector("my_status");
					if (statusNode?.textContent) {
						let malStatus = statusNode.textContent?.toLowerCase();
						if (malStatus === "on-hold") {
							l.status = "HOLD";
						} else if (malStatus === "dropped") {
							l.status = "DROPPED";
						} else if (malStatus === "plan to watch") {
							l.status = "PLANNED";
						} else if (malStatus === "watching") {
							l.status = "WATCHING";
						} else if (malStatus === "completed") {
							l.status = "FINISHED";
						} else {
							console.warn(
								"Anime has no status or an unrecognized status:",
								malStatus,
								"anime_title:",
								titleNode.textContent,
							);
						}
					}
					const typeNode = animeNode.querySelector("series_type");
					if (typeNode?.textContent) {
						const malSeriesType = typeNode.textContent?.toLowerCase();
						if (malSeriesType === "tv" || malSeriesType === "movie") {
							l.type = malSeriesType;
						} else {
							console.warn(
								"Anime has no type or an unrecognized type:",
								malSeriesType,
								"anime_title:",
								titleNode.textContent,
							);
						}
					}
					try {
						const startDateNode = animeNode.querySelector("my_start_date");
						const finishDateNode = animeNode.querySelector("my_finish_date");
						if (
							startDateNode?.textContent &&
							startDateNode?.textContent != "0000-00-00"
						) {
							// For start date, we can simply add the activity manually.
							l.activity = [
								// We don't need all the data when importing activity.
								// customDate must be a date object.
								{
									type: "STATUS_CHANGED",
									data: "WATCHING",
									customDate: new Date(startDateNode.textContent),
								},
								// eslint-disable-next-line @typescript-eslint/no-explicit-any
							] as any[];
						}
						if (
							finishDateNode?.textContent &&
							finishDateNode?.textContent != "0000-00-00"
						) {
							l.datesWatched = [new Date(finishDateNode.textContent)];
						}
					} catch (err) {
						console.error(
							"Processing start/finish times for anime failed!",
							err,
						);
						notify({
							type: "error",
							text: "Failed to process start/finish times for an anime! Check console for more details.",
						});
					}
					rList.push(l);
				}
			} catch (err) {
				console.error("MyAnimeList import failed!", err);
				notify({
					type: "error",
					text: "Failed to process import data!",
				});
			}
		} else if (list?.type === "ryot") {
			importText = "Ryot";
			try {
				const s = JSON.parse(list.data);
				// Builds imported list in previous step for ease.
				rList = s;
			} catch (err) {
				console.error("Ryot import processing failed!", err);
				notify({
					type: "error",
					text: "Processing failed!. Please report this issue if it persists.",
				});
			}
		} else if (list?.type === "todomovies") {
			importText = "TodoMovies";
			try {
				const s = JSON.parse(list.data);
				// Builds imported list in previous step for ease.
				rList = s;
			} catch (err) {
				console.error("TodoMovies import processing failed!", err);
				notify({
					type: "error",
					text: "Processing failed!. Please report this issue if it persists.",
				});
			}
		}
	}

	function addRow(
		ev: FocusEvent & { currentTarget: EventTarget & HTMLInputElement },
	) {
		if (!ev.currentTarget.value) {
			return;
		}
		const lo = { name: ev.currentTarget.value } as ImportedList;
		const yearEl = document.getElementById("addYear") as HTMLInputElement;
		if (yearEl?.value) {
			lo.year = Number(yearEl.value);
		}
		rList.push(lo);
		rList = rList;
		ev.currentTarget.value = "";
		yearEl.value = "";
	}

	function removeRow(l: ImportedList) {
		rList = rList.filter((r) => r.name !== l.name);
		rList = rList;
	}

	/**
	 * inputs that have a `data-validateme` property, we will validate with the
	 * browser.
	 */
	function validatableInputsAreValid(): boolean {
		if (!importTableEl) {
			// Somehow the table isn't defined, i guess allow continuing so we
			// dont block.
			console.warn("validatableInputsAreValid: There is no table element.");
			return true;
		}
		const inputs = importTableEl.querySelectorAll<HTMLInputElement>(
			"tbody input[data-validateme]",
		);
		if (inputs.length <= 0) {
			console.warn("validatableInputsAreValid: No inputs found.");
			return true;
		}
		for (const input of inputs) {
			if (!input.reportValidity()) {
				// Stop the for loop after we hit one input that is invalid.
				// No need to continue after we know at least one is invalid.
				console.error("validatableInputsAreValid: Invalid input found.", input);
				return false;
			}
		}
		return true;
	}

	async function startImport() {
		if (!validatableInputsAreValid()) {
			console.warn(
				"startImport: Some fields are not valid, not starting import!",
			);
			notify({
				type: "error",
				text: "An error was found in one of the inputs, please fix it and try starting the import again.",
			});
			return;
		}
		console.log("startImport: Starting.", rList);
		isImporting = true;
		window.scrollTo(0, 0);
		for (let i = 0; i < rList.length; i++) {
			if (cancelled) {
				notify({ type: "error", text: "Importing Cancelled" });
				return;
			}
			const li = rList[i];
			try {
				console.log("Importing", li);
				await doImport(li);
			} catch (err) {
				li.state = ImportResponseType.IMPORT_FAILED;
				console.error("Failed to import item:", li, "reason:", err);
				notify({
					type: "error",
					text: "Failed to import an item! Check console for more info.",
					time: Infinity,
				});
			}
			await sleep(1500);
		}
		store.importedList = undefined;
		if (
			rList.some(
				(i) =>
					i.state == ImportResponseType.IMPORT_FAILED ||
					i.state == ImportResponseType.IMPORT_NOTFOUND,
			)
		) {
			// Some items failed.. go to some-failed
			store.parsedImportedList = rList;
			goto(resolve("/import/some-failed"));
		} else {
			const ignoredCount = rList.filter(
				(i) => i.state === ImportResponseType.IMPORT_IGNORED,
			).length;
			notify({
				type: "success",
				text:
					ignoredCount > 0
						? `All content imported, apart from ${ignoredCount} you chose to ignore.`
						: "All content successfully imported!",
				time: 15000,
			});
			goto(resolve("/"));
		}
	}

	/**
	 * Import one item.
	 *
	 * `ignoreThisItem` gives up on matching it instead, which imports nothing
	 * and saves that decision, so this name isn't asked about again on any
	 * future import.
	 */
	async function doImport(item: ImportedList, ignoreThisItem = false) {
		if (!item.name?.trim()) {
			item.state = ImportResponseType.IMPORT_NOTFOUND;
			rList = rList;
			return;
		}
		const resp = await req.post<ImportResponse>("/import", {
			...item,
			ignoreSavedMatches,
			ignoreThisItem,
		});
		return new Promise((res, rej) => {
			if (resp.type === ImportResponseType.IMPORT_MULTI) {
				console.log("Import found multiple responses for content", resp);
				let results = resp.results;
				if (!results || results.length <= 0) {
					item.state = ImportResponseType.IMPORT_NOTFOUND;
					rList = rList;
					return;
				}
				if (item.year) {
					results.sort((a, b) => {
						try {
							const ar = a.releaseDate;
							const ay = ar
								? new Date(Date.parse(ar)).getFullYear()
								: undefined;
							const br = b.releaseDate;
							const by = br
								? new Date(Date.parse(br)).getFullYear()
								: undefined;
							if (ay == item.year) return -1;
							else if (by == item.year) return 1;
						} catch (err) {
							console.error("doImport: results sort failed", err);
						}
						return 0;
					});
				}
				importMultiItem = {
					original: item,
					results: results,
					callback: (err) => {
						if (err) {
							item.state = ImportResponseType.IMPORT_NOTFOUND;
							rList = rList;
							rej(err);
						} else {
							res(0);
						}
					},
				};
			} else if (resp.type === ImportResponseType.IMPORT_SUCCESS) {
				item.state = ImportResponseType.IMPORT_SUCCESS;
				const w = resp.watchedEntry;
				if (w) {
					const release = w.media?.releaseDate;
					if (release) item.year = new Date(Date.parse(release)).getFullYear();
					const t = w.media ? getContentTypeFromMedia(w.media) : undefined;
					if (t) item.type = t;
				}
				rList = rList;
				res(0);
			} else {
				item.state = resp.type;
				rList = rList;
				res(0);
			}
		});
	}

	/**
	 * Helper to allow user to quickly update
	 * all statuses to a new one.
	 */
	function changeAllStatuses() {
		changeAllStatusesModalCb = (newStatus) => {
			try {
				console.debug("changeAllStatusesModalCb: newStatus:", newStatus);
				if (!newStatus) {
					// User cancelled flow
					changeAllStatusesModalCb = undefined;
					return;
				}
				if (!rList) {
					console.error("changeAllStatusesModalCb: No list to modify!");
					changeAllStatusesModalCb = undefined;
					return;
				}
				for (let i = 0; i < rList.length; i++) {
					const r = rList[i];
					r.status = newStatus;
				}
				rList = rList;
			} catch (err) {
				console.error("changeAllStatusesModalCb: Failed!", err);
				notify({
					type: "error",
					text: "Failed when updating all statuses. Please try again!",
				});
			}
			changeAllStatusesModalCb = undefined;
		};
	}

	// Not sure why this happens:
	// https://github.com/sveltejs/svelte/discussions/14692#discussioncomment-11569475
	const alist = getList();
</script>

{#await alist}
	<Spinner />
{:then}
	<div class="content">
		<div class="inner">
			{#if rList}
				<h2>Importing {importText ? `From ${importText}` : ""}</h2>
				<h5 class="norm">
					{#if !isImporting}
						Review your imported list and fix any problems.
					{:else}
						You can fix any failed imports when the process completes.
					{/if}
				</h5>
				<div class="table-wrap">
					<table
						bind:this={importTableEl}
						class={isImporting ? "is-importing" : ""}
					>
						<thead>
							<tr>
								{#if isImporting}
									<th class="loading-col"></th>
								{/if}
								<th>Name</th>
								<th>Year</th>
								<th>Type</th>
								<th>Status</th>
								<th>Rating</th>
								{#if !isImporting}
									<th></th>
								{/if}
							</tr>
						</thead>
						<tbody>
							<!-- TODO: Fix this to use a keyed each somehow (need unique id for key) -->
							<!-- eslint-disable-next-line svelte/require-each-key -->
							{#each rList as l}
								<tr>
									{#if isImporting}
										<td class="icon-cell">
											<div>
												{#if !l.state}
													<SpinnerTiny />
												{:else if l.state === ImportResponseType.IMPORT_SUCCESS}
													<Icon i="check" wh={22} />
												{:else if l.state === ImportResponseType.IMPORT_NOTFOUND}
													<Icon i="close" wh={22} />
												{:else if l.state === ImportResponseType.IMPORT_FAILED}
													<Icon i="close" wh={22} />
												{:else if l.state === ImportResponseType.IMPORT_EXISTS}
													<Icon i="check" wh={22} />
												{:else if l.state === ImportResponseType.IMPORT_IGNORED}
													<Icon i="eye-closed" wh={22} />
												{/if}
											</div>
										</td>
									{/if}
									<td class="name">
										<input
											class="plain"
											bind:value={l.name}
											disabled={isImporting}
										/>
									</td>
									<td class="year">
										<input
											class="plain"
											bind:value={l.year}
											placeholder="YYYY"
											type="number"
											disabled={isImporting}
										/>
									</td>
									<td class="type">
										<DropDown
											options={dropDownSupportedTypes}
											bind:active={l.type}
											placeholder="Type"
											blendIn={true}
											disabled={isImporting}
										/>
									</td>
									<td class="status">
										<DropDown
											options={[
												"FINISHED",
												"PLANNED",
												"WATCHING",
												"HOLD",
												"DROPPED",
											]}
											bind:active={l.status}
											placeholder="Status"
											blendIn={true}
											disabled={isImporting}
										/>
									</td>
									<td class="rating">
										<input
											class="plain"
											data-validateme
											bind:value={l.rating}
											placeholder="0"
											type="number"
											min="0"
											max="10"
											step="0.1"
											disabled={isImporting}
										/>
									</td>
									{#if !isImporting}
										<td>
											<button
												class="plain delete"
												onclick={() => {
													removeRow(l);
												}}
											>
												<Icon i="close" wh="25" />
											</button>
										</td>
									{/if}
								</tr>
							{/each}
							{#if !isImporting}
								<tr>
									<td
										><input
											class="plain"
											placeholder="Name"
											onblur={addRow}
										/></td
									>
									<td class="year">
										<input
											class="plain"
											id="addYear"
											placeholder="YYYY"
											type="number"
										/>
									</td>
									<td class="type"></td>
									<td class="status"></td>
									<td class="rating"></td>
									<td></td>
								</tr>
							{/if}
						</tbody>
					</table>
				</div>
				<div class="import-opts">
					<Checkbox
						name="Ask about every match again"
						bind:value={ignoreSavedMatches}
						disabled={isImporting}
					/>
					<span class="opt-desc">
						Ignore matches saved from previous imports, so you can pick again
						for names that were matched incorrectly.
					</span>
				</div>
				<div class="btns">
					<button onclick={() => goto(resolve("/import"))}>
						<Icon i="arrow" />Back
					</button>
					<button onclick={() => changeAllStatuses()} disabled={isImporting}>
						Change All Statuses
					</button>
					<button onclick={startImport} disabled={isImporting}>
						Start Importing
					</button>
				</div>
				{#if typeof changeAllStatusesModalCb === "function"}
					<Modal
						title="Select a New Status"
						desc="Override the status for all content in the table"
						maxWidth="600px"
						onClose={() =>
							typeof changeAllStatusesModalCb === "function"
								? changeAllStatusesModalCb()
								: undefined}
					>
						<Status status={undefined} onChange={changeAllStatusesModalCb} />
					</Modal>
				{/if}
			{:else}
				<h2>No list</h2>
			{/if}
		</div>
	</div>

	<!-- Multiple results found modal -->
	{#if importMultiItem}
		<Modal
			title="Multiple Results Found"
			desc="Select the correct item for {importMultiItem.original
				.name} {importMultiItem.original.year &&
				'(' + importMultiItem.original.year + ')'}"
			onClose={() => {
				importMultiItem?.callback("closed results modal");
				importMultiItem = undefined;
			}}
		>
			<PosterList type="vertical">
				{#each importMultiItem.results as r (r.ids)}
					<Poster
						media={r}
						small={true}
						disableInteraction={true}
						hideButtons={true}
						onClick={async () => {
							const item = rList.find(
								(i) => i.name === importMultiItem?.original.name,
							);
							console.log(
								"MultipleResultsFound: Poster clicked. Item in rList:",
								item,
							);
							if (item) {
								// We found the item in our import list, update it
								// to match the selected choice and do the import with it.
								item.type = getContentTypeFromMedia(r);
								if (item.type === "game") {
									item.igdbId = r.ids.igdb;
								} else if (item.type === "movie" || item.type === "tv") {
									item.tmdbId = r.ids.tmdb;
								} else {
									item.state = ImportResponseType.IMPORT_FAILED;
									notify({
										type: "error",
										text: "Can't import selected result because it has an unsupported type associated with it!",
										time: 10000,
									});
									return;
								}
								try {
									await doImport(item);
									importMultiItem?.callback(undefined);
								} catch (err) {
									importMultiItem?.callback(String(err));
								}
								importMultiItem = undefined;
							} else {
								// TODO: show error notif and update state with error icon
							}
							console.log("multi: Poster clicked", r);
						}}
					/>
				{/each}
			</PosterList>
			<div class="ignore-row">
				<button
					onclick={async () => {
						const item = rList.find(
							(i) => i.name === importMultiItem?.original.name,
						);
						if (!item) {
							return;
						}
						try {
							await doImport(item, true);
							importMultiItem?.callback(undefined);
						} catch (err) {
							importMultiItem?.callback(String(err));
						}
						importMultiItem = undefined;
					}}
				>
					None of these, stop asking about it
				</button>
			</div>
		</Modal>
	{/if}
{:catch err}
	<Error error={err} pretty="Failed to process list!" />
{/await}

<style lang="scss">
	// The give up button sits on its own row under the results, centered so
	// it doesn't read as one of the choices above it.
	.ignore-row {
		display: flex;
		justify-content: center;
		margin-top: 15px;

		button {
			width: max-content;
		}
	}

	.content {
		display: flex;
		width: 100%;
		justify-content: center;
		padding: 0 30px 30px 30px;

		.inner {
			display: flex;
			flex-flow: column;
			min-width: 400px;
			max-width: 1200px;

			@media screen and (max-width: 410px) {
				min-width: 100%;
			}
		}
	}

	.table-wrap {
		overflow: auto;
		margin-top: 20px;
	}

	table {
		td {
			&.name {
				width: 100%;
			}

			&.year {
				width: 70px;
				min-width: 67px;
			}

			&.type {
				width: 120px;
			}

			&.rating {
				width: 82px;
			}
		}
	}

	table.is-importing {
		th {
			padding-left: 3px;
		}

		td {
			padding-left: 3px;

			input {
				padding: 7px 0;

				&:focus {
					padding: 7px 5px;
					padding-left: 3px;
				}
			}
		}
	}

	.import-opts {
		display: flex;
		flex-flow: row;
		align-items: center;
		flex-wrap: wrap;
		margin-top: 20px;
		gap: 8px;

		.opt-desc {
			font-size: 14px;
			opacity: 0.7;
		}
	}

	.btns {
		display: flex;
		flex-flow: row;
		margin-top: 20px;
		gap: 5px;

		button {
			width: max-content;
			gap: 3px;

			&:last-of-type {
				margin-left: auto;
			}
		}
	}
</style>
