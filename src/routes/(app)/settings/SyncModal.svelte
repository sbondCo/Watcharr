<script lang="ts">
	import { t } from "@/lib/i18n/index.svelte";
	// The original JobWatcherModal.svelte component.
	// This could eventually end up using that component or we could keep sync jobs here.

	import Icon from "@/lib/Icon.svelte";
	import Modal from "@/lib/Modal.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import { notify } from "@/lib/util/notify";
	import {
		JobStatus,
		type GetJobResponse,
		type JobCreatedResponse,
	} from "@/types";
	import { onDestroy, onMount } from "svelte";
	import { req } from "@/lib/util/api";

	interface Props {
		type?: "jellyfin" | "plex";
		onClose: () => void;
	}

	let { type = "jellyfin", onClose }: Props = $props();

	let step: "starting" | "errored" | "job-running" | "done" | "modal-closing" =
		$state("starting");
	let jobId: string | undefined;
	let currentTask: string | undefined = $state();
	let latestJobStatus: GetJobResponse | undefined = $state();

	async function startSync() {
		try {
			const r = await req.get<JobCreatedResponse>(
				type === "jellyfin" ? "/jellyfin/sync" : "/plex/sync",
			);
			console.log("startSync: Response:", r);
			if (!r.jobId) {
				step = "errored";
				console.error("startSync: No jobId returned!");
				return;
			}
			jobId = r.jobId;
			step = "job-running";
			startJobWatcher();
		} catch (err) {
			console.error("startSync failed!", err);
			step = "errored";
		}
	}

	async function startJobWatcher() {
		if (!jobId) {
			console.error("startJobWatcher: No Job Id");
			notify({
				text: t("job.noJobId"),
				type: "error",
			});
			return;
		}
		console.log("startJobWatcher: Starting..");
		let seqfailedJobReqs = 0;
		while (step === "job-running") {
			try {
				const r = await req.get<GetJobResponse>(`/job/${jobId}`);
				console.log("jobWatcher: Got job data:", r);
				latestJobStatus = r;
				currentTask = r?.currentTask;
				if (r?.status === JobStatus.DONE) {
					step = "done";
				} else if (r?.status === JobStatus.CANCELLED) {
					step = "errored";
				}
				// If we get here without erroring, we can reset it to 0.
				seqfailedJobReqs = 0;
			} catch (err) {
				console.error(
					"jobWatcher: Get job request failed!",
					seqfailedJobReqs,
					err,
				);
				seqfailedJobReqs++;
			}
			if (seqfailedJobReqs >= 10) {
				console.error("jobWatcher: Failed 10 times in a row!");
				notify({
					text: t("job.statusCheckerFailed"),
					type: "error",
					time: 30000,
				});
				step = "errored";
				break;
			}
			await new Promise((r) => setTimeout(r, 1000));
		}
	}

	function modalClose() {
		if (step === "job-running") {
			notify({
				text: t("job.continuesInBackground"),
				time: 10000,
			});
		}
		step = "modal-closing";
		onClose();
	}

	onMount(() => {
		startSync();
	});

	onDestroy(() => {
		step = "starting";
		jobId = undefined;
		currentTask = undefined;
		latestJobStatus = undefined;
	});
</script>

<Modal
	title={t("sync.title", {
		service:
			type === "jellyfin"
				? localStorage.getItem("useEmby")
					? "Emby"
					: "Jellyfin"
				: "Plex",
	})}
	maxWidth="700px"
	onClose={modalClose}
>
	<div class="ctr">
		{#if step === "done"}
			<Icon i="check" wh={60} />
		{:else if step === "errored"}
			<Icon i="close" wh={70} />
		{:else}
			<Spinner />
		{/if}
		<div>
			{#if step === "starting"}
				<h4 class="norm">{t("job.starting")}</h4>
				<span>{t("sync.requesting")}</span>
			{:else if step === "job-running"}
				<h4 class="norm">{t("sync.syncing")}</h4>
				{#if currentTask}
					<span>{currentTask}</span>
				{/if}
			{:else if step === "done"}
				{#if !latestJobStatus?.errors || latestJobStatus?.errors?.length <= 0}
					<h4 class="norm">{t("job.finished")}</h4>
					<span>{t("sync.finishedNoErrors")}</span>
				{:else}
					<h4 class="norm">
						{t("job.finishedWithErrors", {
							count: latestJobStatus?.errors?.length,
						})}
					</h4>
					<span>{t("sync.finishedButWithErrors")}</span>
					<ul>
						{#each latestJobStatus?.errors as e (e)}
							<li>{e}</li>
						{/each}
					</ul>
				{/if}
			{:else if step === "errored"}
				<h4 class="norm">{t("job.weErrored")}</h4>
				<span>{t("sync.erroredBeforeStart")}</span>
			{:else}
				<h4 class="norm">{t("job.unknownState")}</h4>
				<span>{t("sync.unknownStateDesc")}</span>
			{/if}
		</div>
	</div>
</Modal>

<style lang="scss">
	.ctr {
		display: flex;
		flex-flow: row;
		gap: 20px;
		justify-content: start;
		align-items: start;
		margin-top: 25px;
		margin-bottom: 15px;
		margin-left: 15px;

		& > div:last-of-type {
			display: flex;
			flex-flow: column;
			gap: 8px;
			padding: 8px 0;

			& > span {
				font-style: italic;

				&::first-letter {
					text-transform: uppercase;
				}
			}

			& > ul {
				padding-left: 25px;

				li::first-letter {
					text-transform: uppercase;
				}
			}
		}
	}
</style>
