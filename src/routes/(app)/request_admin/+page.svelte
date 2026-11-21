<script lang="ts">
	import { t } from "@/lib/i18n/index.svelte";
	import { goto } from "$app/navigation";
	import { resolve } from "$app/paths";
	import { req } from "@/lib/util/api";
	import { notify } from "@/lib/util/notify";
	import { store } from "@/store.svelte";
	import { UserPermission } from "@/types";

	let page = $state(0);
	let adminToken: string | undefined = $state();

	function generateAdminToken() {
		req
			.get("/auth/admin_token")
			.then(() => {
				page = 1;
			})
			.catch((err) => {
				console.error("Failed to generate admin token", err);
				notify({ type: "error", text: t("requestAdmin.failedToGenerate") });
			});
	}

	function useAdminToken() {
		if (!adminToken) {
			console.error("useAdminToken: No admin token provided!");
			return;
		}
		req
			.post("/auth/admin_token", { token: adminToken })
			.then(() => {
				notify({ type: "success", text: t("requestAdmin.nowAdmin") });
				if (store.userInfo) {
					store.userInfo.permissions = UserPermission.PERM_ADMIN;
				}
				goto(resolve("/"));
			})
			.catch((err) => {
				console.error("Failed to use admin token", err);
				notify({ type: "error", text: t("requestAdmin.failedToUse") });
			});
	}
</script>

<div class="content">
	<div class="inner">
		<h2>{t("requestAdmin.title")}</h2>

		{#if page == 0}
			<button onclick={generateAdminToken}>{t("request.request")}</button>
		{:else if page == 1}
			<p>
				{t("requestAdmin.checkLog")}
			</p>
			<input
				bind:value={adminToken}
				type="text"
				placeholder={t("requestAdmin.token")}
			/>
			<button onclick={useAdminToken}>{t("requestAdmin.checkToken")}</button>
		{/if}
	</div>
</div>

<style lang="scss">
	.content {
		display: flex;
		width: 100%;
		justify-content: center;
		padding: 0 30px 0 30px;

		.inner {
			display: flex;
			flex-flow: column;
			gap: 10px;
			min-width: 400px;
			max-width: 400px;
			overflow: hidden;
		}
	}
</style>
