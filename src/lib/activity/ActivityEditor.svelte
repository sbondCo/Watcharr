<script lang="ts">
	import { t } from "@/lib/i18n/index.svelte";
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
			error = t("activity.editor.nothingChanged");
			return;
		}

		if (isNaN(date.getTime()) || date.getTime() < 0) {
			error = t("activity.editor.dateInvalid");
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

<Modal
	title={t("activity.editor.title")}
	desc={activityMessage}
	maxWidth="400px"
	{onClose}
>
	{#if error}
		<Error
			pretty={t("activity.editor.updateFailed")}
			{error}
			margin="0 0 10px 0"
		/>
	{/if}

	<div class="centered">
		<h3>{t("activity.editor.date")}</h3>
		<DateInput {date} />

		<h3>{t("activity.editor.time")}</h3>
		<TimeInput {date} />

		<h3>{t("activity.editor.note")}</h3>
		<textarea
			name="note"
			rows="3"
			placeholder={t("activity.editor.notePlaceholder")}
			bind:value={note}></textarea>

		<ActivityEditorRawView {activity} />

		<div class="button-row">
			<button class="danger" onclick={remove}>{t("common.delete")}</button>
			<div>
				<button onclick={update} disabled={!anythingChanged}
					>{t("common.update")}</button
				>
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
