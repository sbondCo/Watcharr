<script lang="ts">
	import { t } from "@/lib/i18n/index.svelte";
	import { store } from "@/store.svelte";
	import Setting from "../settings/Setting.svelte";
	import { RatingStep, RatingSystem } from "@/types";
	import { updateUserSetting } from "../util/api";

	function update(v: RatingSystem) {
		if (!store.userSettings) {
			console.error("No settings.");
			return;
		}
		updateUserSetting("ratingSystem", v, () => {
			store.userSettings!.ratingSystem = v;
		});
	}

	function updateStep(v: number) {
		if (!store.userSettings) {
			console.error("No settings.");
			return;
		}
		updateUserSetting("ratingStep", v, () => {
			store.userSettings!.ratingStep = v;
		});
	}
</script>

<Setting
	title={t("settings.rating.system")}
	desc={t("settings.rating.systemDesc")}
>
	<div class="rat-wrap">
		<button
			class={[
				"plain",
				store.userSettings?.ratingSystem === RatingSystem.OutOf5
					? "active"
					: "",
			].join(" ")}
			onclick={() => update(RatingSystem.OutOf5)}
		>
			0-5
		</button>
		<button
			class={[
				"plain",
				store.userSettings?.ratingSystem === RatingSystem.OutOf10 ||
				!store.userSettings?.ratingSystem
					? "active"
					: "",
			].join(" ")}
			onclick={() => update(RatingSystem.OutOf10)}
		>
			0-10
		</button>
		<button
			class={[
				"plain",
				store.userSettings?.ratingSystem === RatingSystem.OutOf100
					? "active"
					: "",
			].join(" ")}
			onclick={() => update(RatingSystem.OutOf100)}
		>
			0-100
		</button>
		<button
			class={[
				"plain",
				store.userSettings?.ratingSystem === RatingSystem.Thumbs
					? "active"
					: "",
			].join(" ")}
			onclick={() => update(RatingSystem.Thumbs)}
		>
			{t("settings.rating.thumbs")}
		</button>
	</div>
</Setting>

{#if store.userSettings?.ratingSystem === RatingSystem.OutOf10 || store.userSettings?.ratingSystem === RatingSystem.OutOf5 || !store.userSettings?.ratingSystem}
	<Setting
		title={t("settings.rating.step")}
		desc={t("settings.rating.stepDesc")}
	>
		<div class="rat-wrap">
			<button
				class={[
					"plain",
					store.userSettings?.ratingStep === RatingStep.Point1 ? "active" : "",
				].join(" ")}
				onclick={() => updateStep(RatingStep.Point1)}
			>
				0.1
			</button>
			<button
				class={[
					"plain",
					store.userSettings?.ratingStep === RatingStep.Point5 ? "active" : "",
				].join(" ")}
				onclick={() => updateStep(RatingStep.Point5)}
			>
				0.5
			</button>
			<button
				class={[
					"plain",
					store.userSettings?.ratingStep === RatingStep.One ||
					!store.userSettings?.ratingStep
						? "active"
						: "",
				].join(" ")}
				onclick={() => updateStep(RatingStep.One)}
			>
				1
			</button>
		</div>
	</Setting>
{/if}

<style lang="scss">
	.rat-wrap {
		display: flex;
		flex-flow: row;
		/* gap: 5px; */
		border-radius: 10px;
		overflow: auto;

		button {
			display: flex;
			flex-flow: row;
			align-items: center;
			justify-content: center;
			gap: 3px;
			width: 100%;
			padding: 15px 8px;
			color: $text-color;
			background-color: $accent-color;
			/* font-family: "Shrikhand", sans-serif; */
			font-size: 16px;
			transition:
				color 150ms ease-in-out,
				background-color 150ms ease-in-out;

			&:hover,
			&.active {
				color: $bg-color;
				background-color: $accent-color-hover;
				font-weight: bold;
			}

			&:not(:last-of-type) {
				/* border-right: 1px solid $placeholder-color; */
			}

			:global(svg) {
				width: 24px;
			}
		}
	}
</style>
