<script lang="ts">
	import { t } from "@/lib/i18n/index.svelte";
	import Modal from "@/lib/Modal.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import Setting from "@/lib/settings/Setting.svelte";
	import SettingsList from "@/lib/settings/SettingsList.svelte";
	import { toRelativeTime } from "@/lib/util/helpers";
	import { notify } from "@/lib/util/notify";
	import type { AllTasksResponse } from "@/types";
	import { req } from "@/lib/util/api";
	import { onMount } from "svelte";
	import Icon from "@/lib/Icon.svelte";

	interface Props {
		onClose: () => void;
	}

	let { onClose }: Props = $props();

	let now = $state(Date.now());
	let formDisabled = $state(false);
	let taskSchedule: AllTasksResponse[] = $state([]);

	async function getAllTasks() {
		try {
			const res = await req.get<AllTasksResponse[]>("/task/");
			taskSchedule = res?.sort((a, b) => {
				if (a.name < b.name) {
					return -1;
				}
				if (a.name > b.name) {
					return 1;
				}
				return 0;
			});
		} catch (err) {
			console.error("getAllTasks failed!", err);
			notify({
				type: "error",
				text: t("server.tasks.failedToGet"),
				time: 6000,
			});
		}
	}

	async function rescheduleTask(name: string, seconds: number) {
		const nid = notify({ type: "loading", text: t("common.updating") });
		try {
			formDisabled = true;
			const res = await req.putWhole(`/task/${name}`, { seconds });
			if (res.status === 200) {
				notify({ id: nid, type: "success", text: t("server.tasks.updated") });
				getAllTasks();
			} else {
				console.error(
					"rescheduleTask: Unexpected response status code:",
					res.status,
				);
				notify({
					id: nid,
					type: "error",
					text: t("server.tasks.unexpectedResponse"),
					time: 6000,
				});
			}
			formDisabled = false;
		} catch (err) {
			console.error("rescheduleTask failed!", err);
			notify({
				id: nid,
				type: "error",
				text: t("server.tasks.failedToUpdate"),
				time: 6000,
			});
			formDisabled = false;
		}
	}

	async function runTask(name: string) {
		const nid = notify({ type: "loading", text: t("server.tasks.executing") });
		try {
			await req.post(`/task/${name}/run`);
			notify({ id: nid, type: "success", text: t("server.tasks.executed") });
		} catch (err) {
			console.error("runTask failed!", err);
			notify({
				id: nid,
				type: "error",
				text: t("server.tasks.manualRunFailed"),
			});
		}
	}

	onMount(() => {
		getAllTasks();
		const nowInterval = setInterval(() => {
			now = Date.now();
		}, 1000);
		const getTasksInterval = setInterval(() => {
			getAllTasks();
		}, 8000);
		return () => {
			clearInterval(nowInterval);
			clearInterval(getTasksInterval);
		};
	});
</script>

<Modal
	title={t("server.tasks.title")}
	desc={t("server.tasks.desc")}
	maxWidth="700px"
	{onClose}
>
	<SettingsList>
		{#if taskSchedule?.length <= 0}
			<Spinner />
		{:else}
			{#each taskSchedule as task (task.name)}
				{@const nextRun = toRelativeTime(
					(new Date(task.nextRun).getTime() - now) / 1000,
				)}
				<Setting title={task.name}>
					<div class="entry">
						<span>{t("server.tasks.runsEvery")}</span>
						<input
							type="number"
							placeholder="60"
							bind:value={task.seconds}
							disabled={formDisabled}
							onblur={() => {
								rescheduleTask(task.name, task.seconds);
							}}
						/>
						<span>
							{nextRun === t("time.now")
								? t("server.tasks.secondsNextNow")
								: t("server.tasks.secondsNextIn", { time: nextRun })}
						</span>
						<button onclick={() => runTask(task.name)}>
							<Icon i="play" />
						</button>
					</div>
				</Setting>
			{/each}
		{/if}
	</SettingsList>
</Modal>

<style lang="scss">
	.entry {
		display: flex;
		flex-flow: row;
		align-items: center;
		gap: 5px;

		input {
			width: 110px;
		}

		button {
			width: min-content;
			height: 100%;
			margin-left: auto;
		}
	}
</style>
