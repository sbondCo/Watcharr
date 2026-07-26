<script lang="ts">
	import Modal from "@/lib/Modal.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import Setting from "@/lib/settings/Setting.svelte";
	import SettingsList from "@/lib/settings/SettingsList.svelte";
	import { toRelativeTime } from "@/lib/util/helpers";
	import { notify } from "@/lib/util/notify";
	import type { AllTasksResponse } from "@/types";
	import { req } from "@/lib/util/api";
	import { onMount } from "svelte";

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
				text: "Failed to get all tasks from server.",
				time: 6000,
			});
		}
	}

	async function rescheduleTask(name: string, seconds: number) {
		const nid = notify({ type: "loading", text: "Updating.." });
		try {
			formDisabled = true;
			const res = await req.putWhole(`/task/${name}`, { seconds });
			if (res.status === 200) {
				notify({ id: nid, type: "success", text: "Schedule updated." });
				getAllTasks();
			} else {
				console.error(
					"rescheduleTask: Unexpected response status code:",
					res.status,
				);
				notify({
					id: nid,
					type: "error",
					text: "Unexpected response from reschedule request.",
					time: 6000,
				});
			}
			formDisabled = false;
		} catch (err) {
			console.error("rescheduleTask failed!", err);
			notify({
				id: nid,
				type: "error",
				text: "Failed to update schedule.",
				time: 6000,
			});
			formDisabled = false;
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
	title="Tasks Schedule"
	desc="Want a routine task to occur more or less frequently? Configure it below."
	{onClose}
>
	<SettingsList>
		{#if taskSchedule?.length <= 0}
			<Spinner />
		{:else}
			{#each taskSchedule as task}
				{@const nextRun = toRelativeTime(
					(new Date(task.nextRun).getTime() - now) / 1000,
				)}
				<Setting title={task.name}>
					Runs every&nbsp;
					<input
						type="number"
						placeholder="60"
						bind:value={task.seconds}
						disabled={formDisabled}
						onblur={() => {
							rescheduleTask(task.name, task.seconds);
						}}
					/>
					&nbsp;seconds. Next{nextRun === "now" ? "" : " in"}
					{nextRun}.
				</Setting>
			{/each}
		{/if}
	</SettingsList>
</Modal>

<style lang="scss">
	input {
		width: 200px;
		margin-top: 5px;
	}
</style>
