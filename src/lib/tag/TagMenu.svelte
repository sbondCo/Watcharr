<script lang="ts">
	import { store } from "@/store.svelte";
	import Icon from "../Icon.svelte";
	import CreateTagModal from "./CreateTagModal.svelte";
	import type { Tag as TagT } from "@/types";
	import Tag from "./Tag.svelte";
	import DeleteTagModal from "./DeleteTagModal.svelte";
	import Menu, { type MenuConfig } from "../Menu.svelte";

	interface Props {
		titleText?: string | undefined;
		onTagClick?: (tag: TagT, remove: boolean) => void | undefined;
		selectedTags?: TagT[] | undefined;
		/**
		 * When `showManageBtn` is true, a manage icon will appear at top
		 * of menu for the user to click. When toggled on, clicking a tag
		 * will trigger a deletion instead of `onTagClick()`.
		 */
		showManageBtn?: boolean;
		menuConfig?: MenuConfig;
		clickOutsideCallback?: () => void;
	}

	const defaultMenuConfig = {
		width: "200px",
		right: "47px",
		arrowLeft: "78px",
	};

	let {
		titleText = undefined,
		onTagClick = undefined!,
		selectedTags = undefined,
		showManageBtn = false,
		menuConfig = {},
		clickOutsideCallback,
	}: Props = $props();

	let allTags = $derived(store.tags);

	let tagModalOpen = $state(false);
	let inManageMode = $state(false);
	let tagToDelete: TagT | undefined = $state(undefined);

	function deleteTag(t: TagT) {
		// This will show the DeleteTagModal (look below).
		tagToDelete = t;
	}
</script>

<Menu
	{clickOutsideCallback}
	conf={Object.assign(defaultMenuConfig, menuConfig)}
>
	<div class="title">
		<h4 class="norm sm-caps">{titleText ? titleText : "my tags"}</h4>
		{#if showManageBtn}
			<button
				class={["plain", inManageMode ? "manage-on" : ""].join(" ")}
				onclick={() => (inManageMode = !inManageMode)}
			>
				<Icon i="trash" wh={18} />
			</button>
		{/if}
		<button class="plain" onclick={() => (tagModalOpen = !tagModalOpen)}>
			<Icon i="add" wh={22} />
		</button>
	</div>
	{#if allTags && allTags.length > 0}
		{#if inManageMode}
			<strong style="font-size: 12px; margin-bottom: 10px;"
				>Click a tag to delete it.</strong
			>
		{/if}
		<div class="list">
			{#each allTags as t}
				{@const isSelected = selectedTags
					? selectedTags.find((tag) => tag.id === t.id)
						? true
						: false
					: false}
				<div>
					<Tag
						tag={t}
						onClick={() =>
							inManageMode ? deleteTag(t) : onTagClick(t, isSelected)}
					/>
					{#if isSelected}
						<Icon i="check" wh={18} />
					{/if}
				</div>
			{/each}
		</div>
	{:else}
		<span style="margin-top: 0;">You have no tags yet!</span>
	{/if}
</Menu>

{#if tagModalOpen}
	<CreateTagModal onClose={() => (tagModalOpen = false)} />
{/if}

{#if tagToDelete}
	<DeleteTagModal tag={tagToDelete} onClose={() => (tagToDelete = undefined)} />
{/if}

<style lang="scss">
	h4 {
		color: $text-color;

		&:not(:first-of-type) {
			margin-top: 8px;
		}
	}

	.title {
		display: flex;
		flex-flow: row;
		align-items: center;
		margin-bottom: 8px;
		gap: 5px;

		button.plain {
			display: flex;
			align-items: center;
			justify-content: center;
			width: 28px;
			height: 26px;
			padding: 2px 3px;
			border-radius: 8px;

			&.manage-on {
				color: #f3555a;
				background-color: $text-color;
			}

			&:first-of-type {
				margin-left: auto;
			}
		}
	}

	.list {
		display: flex;
		flex-flow: column;
		gap: 5px;

		& > div {
			display: flex;
			align-items: center;
			gap: 5px;
			color: $text-color;

			:global(svg) {
				min-width: 18px;
			}
		}
	}
</style>
