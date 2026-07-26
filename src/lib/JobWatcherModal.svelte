<script lang="ts">
	import Icon from "@/lib/Icon.svelte";
	import Modal from "@/lib/Modal.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import { notify } from "@/lib/util/notify";
	import { JobStatus, type GetJobResponse } from "@/types";
	import { onDestroy, onMount } from "svelte";
	import { req } from "./util/api";
	import { ReqerError } from "./util/fetch";

	interface Props {
		modalTitle: string;
		// Promise that returns the job id to watch.
		getJobId: () => Promise<{ jobId: string } | undefined>;
		onClose: () => void;
		messages: {
			starting: string;
		};
	}

	let { modalTitle, getJobId, onClose, messages }: Props = $props();

	let step: "starting" | "errored" | "job-running" | "done" | "modal-closing" =
		$state("starting");
	let jobId: string | undefined;
	let currentTask: string | undefined = $state();
	let latestJobStatus: GetJobResponse | undefined = $state();
	let jobFailError: string | undefined = $state();

	async function startSync() {
		try {
			const r = await getJobId();
			console.log("startSync: Response:", r);
			if (!r?.jobId) {
				step = "errored";
				console.error("startSync: No jobId returned!");
				return;
			}
			jobId = r.jobId;
			step = "job-running";
			startJobWatcher();
		} catch (err: any) {
			console.error("startSync failed!", err);
			step = "errored";
			jobFailError = ReqerError.getMsg(err, "Starting sync failed");
		}
	}

	async function startJobWatcher() {
		if (!jobId) {
			console.error("startJobWatcher: No Job Id");
			notify({
				text: "Unable to start job watcher, no job id.",
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
					text: "Status checker has failed 10 times in a row!",
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
				text: "Sync will continue in the background.. please refresh the page periodically to view your updated list or come back later.",
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

<Modal title={modalTitle} maxWidth="700px" onClose={modalClose}>
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
				<h4 class="norm">Starting</h4>
				<span>{messages.starting}</span>
			{:else if step === "job-running"}
				<h4 class="norm">Running</h4>
				{#if currentTask}
					<span>{currentTask}</span>
				{/if}
			{:else if step === "done"}
				{#if !latestJobStatus?.errors || latestJobStatus?.errors?.length <= 0}
					<h4 class="norm">Finished</h4>
					<span>We have finished. Looks like there were no errors!</span>
				{:else}
					<h4 class="norm">
						Finished With {latestJobStatus?.errors?.length} Error{latestJobStatus
							?.errors?.length === 1
							? ""
							: "s"}
					</h4>
					<span>Job finished, but with errors:</span>
					<ul>
						{#each latestJobStatus?.errors as e}
							<li>{e}</li>
						{/each}
					</ul>
				{/if}
			{:else if step === "errored"}
				<h4 class="norm">We Errored!</h4>
				{#if jobFailError}
					<span>{jobFailError}</span>
				{:else}
					<span
						>We errored before starting the job or the job was cancelled.</span
					>
				{/if}
				{#if latestJobStatus?.errors && latestJobStatus?.errors?.length > 0}
					<ul>
						{#each latestJobStatus?.errors as e}
							<li>{e}</li>
						{/each}
					</ul>
				{/if}
			{:else}
				<h4 class="norm">Unknown State!</h4>
				<span>We're not sure of the current job status.</span>
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
