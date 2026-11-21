<script lang="ts">
	import { t } from "@/lib/i18n/index.svelte";
	import { goto } from "$app/navigation";
	import type { AuthResponse, AvailableAuthProviders } from "@/types";
	import { onMount } from "svelte";
	import { notify, unNotify } from "@/lib/util/notify";
	import { noAuthReq } from "@/lib/util/api";
	import { ReqerError } from "@/lib/util/fetch";
	import { resolve } from "$app/paths";

	let error: string | undefined = $state();

	onMount(() => {
		if (localStorage.getItem("token")) {
			goto(resolve("/"));
		}

		noAuthReq.get<AvailableAuthProviders>("/auth/available").then((r) => {
			if (r) {
				if (!r.isInSetup) {
					console.log("Server not in setup.. navigating to login page.");
					goto(resolve("/(plain)/login"));
				}
			}
		});
	});

	function handleLogin(ev: SubmitEvent) {
		ev.preventDefault();
		const fd = new FormData(ev.target! as HTMLFormElement);
		const user = fd.get("username");
		const pass = fd.get("password");

		if (!user || !pass) {
			error = t("login.usernameAndPasswordRequired");
			return;
		}

		const nid = notify({ text: t("setup.settingUpAdmin"), type: "loading" });
		noAuthReq
			.post<AuthResponse>("/setup/create_admin", {
				username: user,
				password: pass,
			})
			.then((resp) => {
				if (resp.token) {
					console.log("Received token... logging in.");
					localStorage.setItem("token", resp.token);
					goto(resolve("/"));
					notify({
						id: nid,
						text: t("login.welcomeUser", { user }),
						type: "success",
					});
				}
			})
			.catch((err) => {
				error = ReqerError.getMsg(err, t("setup.settingUpFailed"));
				unNotify(nid);
			});
	}
</script>

<svelte:head>
	<title>{t("setup.pageTitle")}</title>
</svelte:head>

<div>
	<div class="inner">
		<div class="headers">
			<h2>{t("setup.title")}</h2>
			<h5 class="norm">{t("setup.desc")}</h5>
		</div>

		{#if error}
			<span class="error">{error}!</span>
		{/if}

		<form onsubmit={handleLogin}>
			<label for="username">{t("login.username")}</label>
			<input type="text" name="username" placeholder={t("login.username")} />

			<label for="password">{t("login.password")}</label>
			<input
				type="password"
				name="password"
				placeholder={t("login.password")}
			/>

			<div class="login-btns">
				<button type="submit"
					><span class="watcharr">W</span>{t("setup.setUp")}</button
				>
			</div>
		</form>
	</div>
</div>

<style lang="scss">
	div,
	form {
		display: flex;
		flex-flow: column;
		align-items: center;
		gap: 10px;
		margin: 0 35px;
	}

	.inner,
	form {
		width: 100%;
		max-width: 400px;
	}

	.inner h2 {
		font-weight: normal;
	}

	.headers {
		display: flex;
		flex-flow: column;
		width: 100%;
		gap: 0;
		margin-bottom: 10px;
	}

	label {
		align-self: flex-start;
		font-weight: bold;
	}

	.login-btns {
		display: flex;
		flex-flow: row;
		gap: 10px;
		width: 100%;

		button {
			display: flex;
			flex-flow: row;
			gap: 10px;

			.watcharr {
				font-family: "Rampart One";
				font-size: 18px;
				line-height: 18px;
			}
		}
	}

	.error {
		display: flex;
		justify-content: center;
		width: 100%;
		padding: 10px;
		background-color: rgb(221, 48, 48);
		text-transform: capitalize;
		color: white;
	}
</style>
