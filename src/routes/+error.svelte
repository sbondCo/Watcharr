<script lang="ts">
	import { t } from "@/lib/i18n/index.svelte";
	import { goto } from "$app/navigation";
	import { page } from "$app/state";
	import { resolve } from "$app/paths";
</script>

<div>
	<div>
		<h1>{page.status}</h1>
		{#if page.status === 404}
			<h4 class="norm">{t("errorPage.notFound")}</h4>
		{:else if page.status === 500}
			<h4 class="norm">{t("errorPage.serverError")}</h4>
		{:else if page?.error?.message}
			<h4 class="norm">{page.error.message}</h4>
		{:else}
			<h4 class="norm">{t("errorPage.couldNotLoad")}</h4>
		{/if}
		<div class="btns">
			<button onclick={() => goto(resolve("/"))}>{t("errorPage.home")}</button>
			<button onclick={() => location.reload()}>{t("errorPage.refresh")}</button
			>
		</div>
	</div>
</div>

<style lang="scss">
	div {
		display: flex;
		flex-flow: column;
		align-items: center;
		justify-content: center;
		height: 100dvh;

		& > div {
			display: flex;
			flex-flow: column;
			max-width: 500px;
			height: max-content;
			align-items: center;
			justify-content: center;
			gap: 10px;
		}

		h1 {
			line-height: 23px;
		}

		&.btns {
			display: flex;
			flex-flow: row;
			gap: 5px;
			margin-top: 10px;
		}
	}
</style>
