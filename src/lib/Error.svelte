<script lang="ts">
	interface Props {
		pretty: string;
		error: any;
		onRetry?: () => void | undefined;
	}

	let { pretty, error, onRetry = undefined }: Props = $props();
</script>

<div class="error-wrap">
	<div>
		<div class="error-text">
			<strong>{pretty}</strong>
			{#if error?.message}
				<p>{error.message}</p>
				{#if error.response?.data?.error}
					<p>{error.response.data.error}</p>
				{/if}
			{:else}
				<p>{JSON.stringify(error)}</p>
			{/if}
		</div>
		<div>
			{#if onRetry}
				<button onclick={onRetry}>Try Again</button>
			{/if}
		</div>
	</div>
</div>

<style lang="scss">
	div.error-wrap {
		display: flex;
		justify-content: center;
		width: 100%;

		& > div {
			display: flex;
			justify-content: center;
			flex-flow: row wrap;
			width: 100%;
			max-width: 500px;
			gap: 5px;
			border-radius: 8px;
			padding: 10px;
			margin: 15px 5px;
			color: white;
			background-color: $error;

			& > div {
				display: flex;
				justify-content: center;
				flex-flow: column;
				width: max-content;

				&:first-of-type {
					margin-right: auto;
				}

				&:last-of-type {
					margin-left: auto;
				}
			}

			button {
				width: fit-content;
			}
		}
	}
</style>
