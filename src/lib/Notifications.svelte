<script lang="ts">
	import { store } from "@/store.svelte";
	import SpinnerTiny from "./SpinnerTiny.svelte";
	import { unNotify } from "./util/notify";
	import Icon from "./Icon.svelte";
</script>

<div id="notifications">
	{#each store.notifications as n (n.id)}
		<div class={`${n.type} notif`}>
			{#if n.type === "loading"}
				<SpinnerTiny />
			{/if}
			<!-- only comes from our strings (which may have html) -->
			<!-- eslint-disable-next-line -->
			<span>{@html n.text}</span>
			<button
				class="plain"
				onclick={() => {
					unNotify(n.id);
				}}
			>
				<Icon i="close" />
			</button>
		</div>
	{/each}
</div>

<style lang="scss">
	#notifications {
		display: flex;
		flex-flow: column;
		gap: 10px;
		position: fixed;
		bottom: 0;
		left: 50%;
		transform: translateX(-50%);
		margin-bottom: 8px;
		z-index: 99999;

		.notif {
			display: flex;
			flex-flow: row;
			align-items: center;
			min-width: 200px;
			color: black;
			background-color: white;
			border-radius: 8px;
			border: 1px solid rgba($color: #000000, $alpha: 0.2);
			box-shadow: 0 4px 10px rgba($color: #000000, $alpha: 0.2);
			animation: comein 250ms ease forwards;
			position: relative;

			&.loading {
				padding-left: 10px;
			}

			@keyframes comein {
				from {
					opacity: 0;
				}

				to {
					opacity: 1;
				}
			}

			&.error {
				color: white;
				background-color: $error;
				border: 1px solid $error;

				span {
					border-color: rgba($color: white, $alpha: 0.5);
				}
			}

			&.success {
				color: white;
				background-color: $success;
				border: 1px solid $success;

				span {
					border-color: rgba($color: white, $alpha: 0.5);
				}
			}

			span {
				width: 100%;
				height: 100%;
				padding-right: 12px;
				border-right: 1px solid rgba($color: black, $alpha: 0.2);
				padding: 10px 12px;
				padding-left: 9px;

				a {
					color: white;
					text-decoration: underline;
				}
			}

			button {
				display: flex;
				align-items: center;
				margin: 8px;
				width: 22px;
				height: 100%;
				color: inherit;
			}
		}
	}
</style>
