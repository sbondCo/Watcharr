<script lang="ts">
	import { goto } from "$app/navigation";
	import { resolve } from "$app/paths";
	import { page } from "$app/state";
	import tooltip from "../actions/tooltip";
	import dismissable from "../attachments/dismissable.svelte";
	import Icon from "../Icon.svelte";
	import TagMenu from "../tag/TagMenu.svelte";
	import DetailedMenu from "./DetailedMenu.svelte";
	import FaceMenu from "./FaceMenu.svelte";
	import FollowingMenu from "./FollowingMenu.svelte";

	let faceMenuShown = $state(false);
	let faceMenuBtn: HTMLButtonElement | undefined = $state();
	let followingMenuShown = $state(false);
	let followingMenuBtn: HTMLButtonElement | undefined = $state();
	let detailedMenuShown = $state(false);
	let detailedMenuBtn: HTMLButtonElement | undefined = $state();
	let tagMenuShown = $state(false);
	let tagMenuBtn: HTMLButtonElement | undefined = $state();
</script>

<div class="ctr">
	<!-- EXTRA DETAILS (only supported for some pages) -->
	{#if page.url?.pathname === "/" || page.url?.pathname.startsWith("/search")}
		<div
			{@attach detailedMenuShown &&
				dismissable(() => (detailedMenuShown = false))}
		>
			<button
				class="plain other detailedView"
				onclick={() => {
					detailedMenuShown = !detailedMenuShown;
				}}
				use:tooltip={{
					text: "Detailed View",
					pos: "bot",
					condition: !detailedMenuShown,
				}}
				bind:this={detailedMenuBtn}
			>
				<Icon i="eye" />
			</button>
			{#if detailedMenuShown}
				<DetailedMenu trackEl={detailedMenuBtn} />
			{/if}
		</div>
	{/if}

	<!-- TAGS -->
	<div {@attach tagMenuShown && dismissable(() => (tagMenuShown = false))}>
		<button
			class="plain other tag"
			onclick={() => {
				tagMenuShown = !tagMenuShown;
			}}
			use:tooltip={{ text: "Tags", pos: "bot", condition: !tagMenuShown }}
			bind:this={tagMenuBtn}
		>
			<Icon i="tag" />
		</button>
		{#if tagMenuShown}
			<TagMenu
				onTagClick={(tag) => {
					goto(resolve(`/tag/${tag.id}`));
					tagMenuShown = false;
				}}
				showManageBtn={true}
				trackEl={tagMenuBtn}
			/>
		{/if}
	</div>

	<!-- DISCOVER -->
	<div>
		<button
			class="plain other discover"
			onclick={() => goto(resolve("/discover"))}
			use:tooltip={{ text: "Discover", pos: "bot" }}
		>
			<Icon i="compass" wh={26} />
		</button>
	</div>

	<!-- FOLLOWING -->
	<div
		{@attach followingMenuShown &&
			dismissable(() => (followingMenuShown = false))}
	>
		<button
			class="plain other following"
			onclick={() => {
				followingMenuShown = !followingMenuShown;
			}}
			use:tooltip={{
				text: "Following",
				pos: "bot",
				condition: !followingMenuShown,
			}}
			bind:this={followingMenuBtn}
		>
			<Icon i="people" wh={26} />
		</button>
		{#if followingMenuShown}
			<FollowingMenu
				close={() => (followingMenuShown = false)}
				trackEl={followingMenuBtn}
			/>
		{/if}
	</div>

	<!-- FACE -->
	<div {@attach faceMenuShown && dismissable(() => (faceMenuShown = false))}>
		<button
			class="plain face"
			onclick={() => {
				faceMenuShown = !faceMenuShown;
			}}
			bind:this={faceMenuBtn}>:)</button
		>
		{#if faceMenuShown}
			<FaceMenu trackEl={faceMenuBtn} />
		{/if}
	</div>
</div>

<style lang="scss">
	.ctr {
		display: flex;
		flex-flow: row;
		justify-content: end;
		gap: 10px;
		align-items: center;

		& > div {
			width: 30px;
			height: 30px;
			display: flex;
			align-items: center;
			justify-content: center;

			button {
				display: flex;
				align-items: center;
				justify-content: center;
				width: 30px;
				height: 30px;
			}
		}

		button.other {
			padding-top: 2px;
			width: 28px;
			transition:
				fill 150ms ease,
				stroke 150ms ease,
				stroke-width 150ms ease;
			fill: $text-color;

			&:hover,
			&:focus-visible {
				:global(path) {
					fill: none;
					stroke: $text-color;
					stroke-width: 30px;
					stroke-linejoin: round;
				}
			}
		}

		button.discover {
			transition:
				fill 150ms ease,
				stroke 150ms ease,
				stroke-width 150ms ease,
				transform 150ms ease;

			&:hover,
			&:focus-visible {
				transform: rotate(60deg);
			}
		}

		button.face {
			padding-left: 4px;
			font-family:
				"Shrikhand",
				system-ui,
				-apple-system,
				BlinkMacSystemFont;
			font-size: 25px;
			transform: rotate(90deg);
			cursor: pointer;
			margin-left: 3px;
			transition:
				-webkit-text-stroke 150ms ease,
				color 150ms ease;

			&:hover,
			&:focus-visible {
				color: $bg-color;
				-webkit-text-stroke: 1.5px $text-color;
			}
		}
	}
</style>
