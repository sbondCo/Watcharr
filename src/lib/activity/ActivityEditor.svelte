<script lang="ts">
	import { updateActivity, removeActivity } from "@/lib/util/api";
	import Modal from "../Modal.svelte";
	import type { Activity } from "@/types";
	import ActivityEditorRawView from "./ActivityEditorRawView.svelte";
	import DateInput from "../input/DateInput.svelte";
	import { SvelteDate } from "svelte/reactivity";
	import TimeInput from "../input/TimeInput.svelte";
	import Error from "../Error.svelte";

	interface Props {
		activity: Activity;
		activityMessage: string;
		onClose: () => void;
		onRemoved: (activity: Activity) => void;
		onUpdated: (activityId: number, updatedActivity: Activity) => void;
	}

	let { activity, activityMessage, onClose, onRemoved, onUpdated }: Props =
		$props();

	let initialDate = Date.parse(activity.customDate ?? activity.createdAt);

	let error: string | undefined = $state();
	let date = new SvelteDate(initialDate);
	let note = $state(activity.note);
	let anythingChanged: boolean = $derived(
		date.getTime() != initialDate || note != activity.note,
	);

	async function update() {
		error = "";

		if (!anythingChanged) {
			error = "Nothing changed!";
			return;
		}

		if (isNaN(date.getTime()) || date.getTime() < 0) {
			error = "Date is invalid!";
			return;
		}

		const updatedActivity = await updateActivity(activity, {
			customDate: date,
			note: note,
		});
		if (!updatedActivity) {
			// Failed..
			return;
		}
		onUpdated(updatedActivity.id, updatedActivity);
		onClose();
		return;
	}

	async function remove() {
		error = "";
		const success = await removeActivity(activity.id);
		if (!success) {
			return;
		}
		onRemoved(activity);
		onClose();
	}
</script>

<Modal title="Edit Activity" desc={activityMessage} maxWidth="400px" {onClose}>
	{#if error}
		<Error pretty="Update Failed" {error} margin="0 0 10px 0" />
	{/if}

	<div class="centered">
		<h3>Date</h3>
		<DateInput {date} />

		<h3>Time</h3>
		<TimeInput {date} />

		<h3>Note</h3>
		<textarea
			name="note"
			rows="3"
			placeholder="Reasons, thoughts or whatever."
			bind:value={note}></textarea>

		<ActivityEditorRawView {activity} />

		<div class="button-row">
			<button class="danger" onclick={remove}>Delete</button>
			<div>
				<button onclick={update} disabled={!anythingChanged}>Update</button>
			</div>
		</div>
	</div>
</Modal>

<style lang="scss">
	.centered {
		display: flex;
		flex-flow: column;
		gap: 10px;
		height: 100%;

		h3 {
			font-size: 16px;
			font-family:
				sans-serif,
				system-ui,
				-apple-system,
				BlinkMacSystemFont;
		}

		.button-row {
			display: flex;
			flex-flow: row;
			justify-content: space-between;
			margin-top: 10px;

			button {
				margin-top: auto;
				width: max-content;
			}
		}
	}
</style>
