<script lang="ts">
	import { t } from "@/lib/i18n/index.svelte";
	import { goto } from "$app/navigation";
	import { page } from "$app/state";
	import Icon from "@/lib/Icon.svelte";
	import { type AuthResponse, type AvailableAuthProviders } from "@/types";
	import { noAuthReq } from "@/lib/util/api";
	import { onMount } from "svelte";
	import { notify, unNotify } from "@/lib/util/notify";
	import { ReqerError } from "@/lib/util/fetch";
	import { resolve } from "$app/paths";

	let error: string | undefined = $state();
	let login = $state(true);
	let availableProviders: string[] = $state([]);
	let apHeader = $state(false);
	let apPlex = $state(false);
	let signupEnabled = $state(true);
	let useEmby = $state(false);
	let noAuto = $state(false);

	onMount(() => {
		if (localStorage.getItem("token")) {
			goto(resolve("/"));
		}

		if (!error && page.url.searchParams.get("again")) {
			error = t("login.loginAgain");
		}
		if (page.url.searchParams.get("noAuto") == "1") {
			console.info(
				"login: Found noAuto param.. auto logins should be disabled now.",
			);
			noAuto = true;
		}

		processAvailableLoginMethods();
	});

	/**
	 * Get and process available login methods.
	 */
	async function processAvailableLoginMethods() {
		let r: AvailableAuthProviders;
		try {
			r = await noAuthReq.get<AvailableAuthProviders>("/auth/available");
		} catch {
			notify({
				type: "error",
				time: 10000,
				text: t("login.failedToGetMethods"),
			});
			return;
		}
		if (!r) {
			console.log("AvailableAuth: no response data");
		}
		if (r.isInSetup) {
			console.log(
				"AvailableAuth: Server is in setup.. navigating to web setup page.",
			);
			goto(resolve("/setup"));
		}
		availableProviders = r.available;
		apHeader = availableProviders?.includes("header");
		apPlex = availableProviders?.includes("plex");
		signupEnabled = r.signupEnabled;
		useEmby = r.useEmby;
		if (r.headerAuthAutoLogin && !noAuto) {
			console.log(
				"AvailableAuth: handling headerAuthAutoLogin.. calling proxyLogin automatically now.",
			);
			proxyLogin(true);
		}
	}

	function handleLogin(ev: SubmitEvent) {
		ev.preventDefault();
		const fd = new FormData(ev.target! as HTMLFormElement);
		const user = fd.get("username");
		const pass = fd.get("password") ?? "";
		const isJellyfin = (ev.submitter as HTMLButtonElement)?.name === "jellyfin";

		if (isJellyfin) {
			if (!user) {
				error = t("login.usernameRequired");
				return;
			}
		} else {
			if (!user || !pass) {
				error = t("login.usernameAndPasswordRequired");
				return;
			}
		}

		let customAuthEP = "";
		if (isJellyfin) {
			customAuthEP = "jellyfin";
		}

		const nid = notify({ text: t("login.loggingIn"), type: "loading" });
		noAuthReq
			.post<AuthResponse>(`/auth${login ? `/${customAuthEP}` : "/register"}`, {
				username: user,
				password: pass,
			})
			.then((resp) => {
				if (resp?.token) {
					console.log("Received token... logging in.");
					localStorage.setItem("token", resp.token);
					if (useEmby) {
						localStorage.setItem("useEmby", "1");
					} else {
						localStorage.removeItem("useEmby");
					}
					goto(resolve("/"));
					notify({
						id: nid,
						text: t("login.welcomeUser", { user }),
						type: "success",
					});
				}
			})
			.catch((err) => {
				error = ReqerError.getMsg(err, t("login.loginFailed"));
				unNotify(nid);
			});
	}

	async function plexLogin() {
		try {
			const { preparePlexAuth, doPlexLogin, plexPinPoll } =
				await import("@/lib/util/plex");
			const p = preparePlexAuth();
			const pin = await doPlexLogin(p);
			plexPinPoll(pin, p, (err, token) => {
				if (err) {
					error = t("login.plexAuthFailed");
					console.error("Plex auth failed!", err);
					return;
				}
				const nid = notify({ text: t("login.loggingIn"), type: "loading" });
				noAuthReq
					.post<AuthResponse>("/auth/plex", {
						token,
						clientIdentifier: p.clientId,
					})
					.then((resp) => {
						if (resp?.token) {
							console.log("Received token... logging in.");
							localStorage.setItem("token", resp.token);
							goto(resolve("/"));
							notify({ id: nid, text: t("login.welcome"), type: "success" });
						}
					})
					.catch((err) => {
						console.error("plexLogin: Fail", err);
						error = ReqerError.getMsg(err, t("login.loginFailed"));
						notify({ id: nid, text: t("common.failed"), type: "error" });
					});
			});
		} catch (err) {
			console.error("plexLogin: failed!", err);
			error = t("login.plexLoginFailed");
		}
	}

	function proxyLogin(auto = false) {
		const nid = notify({ text: t("login.loggingIn"), type: "loading" });
		noAuthReq
			.post<AuthResponse>(`/auth/proxy`)
			.then((resp) => {
				if (resp?.token) {
					console.log("Received token... logging in.");
					localStorage.setItem("token", resp.token);
					goto(resolve("/"));
					notify({ id: nid, text: t("login.welcome"), type: "success" });
				}
			})
			.catch((err) => {
				error = ReqerError.getMsg(err, t("login.loginFailed"));
				if (auto) {
					notify({
						id: nid,
						text: t("login.automaticSsoFailed"),
						type: "error",
					});
				} else {
					unNotify(nid);
				}
			});
	}
</script>

<div>
	<div class="inner">
		<h2>
			{#if login}
				{t("login.getBackIn")}
			{:else}
				{t("login.luckyYouFoundUs")}
			{/if}
		</h2>

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

			{#if login}
				<span class="login-with" style="font-weight: bold"
					>{t("login.loginWith")}</span
				>
				<div class="login-btns">
					<button type="submit"><span class="watcharr">W</span>Watcharr</button>
					{#if availableProviders?.length > 0}
						{#if availableProviders.find((ap) => ap === "jellyfin")}
							{#if useEmby}
								<button type="submit" name="jellyfin" class="other">
									<Icon i="emby" wh={18} />
									emby
								</button>
							{:else}
								<button type="submit" name="jellyfin" class="other">
									<Icon i="jellyfin" wh={18} />
									jellyfin
								</button>
							{/if}
						{/if}
					{/if}
				</div>
				{#if apHeader || apPlex}
					<p style="font-weight: bold; font-size: 14px;">{t("login.or")}</p>
					{#if apHeader}
						<div class="login-btns">
							<button
								type="button"
								name="proxy"
								class="proxy other"
								onclick={() => {
									proxyLogin();
								}}
							>
								<Icon i="lock-closed" wh={18} />{t("login.continueWithSso")}
							</button>
						</div>
					{/if}
					{#if apPlex}
						<div class="login-btns">
							<button
								type="button"
								onclick={() => {
									plexLogin();
								}}
								name="plex"
								class="plex other"
							>
								<Icon i="plex" wh={18} />{t("login.continueWith", {
									service: "Plex",
								})}
							</button>
						</div>
					{/if}
				{/if}
			{:else}
				<div class="login-btns">
					<button type="submit">{t("login.signUp")}</button>
				</div>
			{/if}
		</form>

		{#if signupEnabled}
			<button
				class="plain"
				onclick={() => {
					login = !login;
				}}
			>
				{#if login}
					{t("login.notAUser")}
				{:else}
					{t("login.alreadyAUser")}
				{/if}
			</button>
		{/if}
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

	label {
		align-self: flex-start;
		font-weight: bold;
	}

	span.login-with {
		font-size: 14px;
	}

	.login-btns {
		display: flex;
		flex-flow: row;
		gap: 10px;
		width: 100%;

		/* Hardcoded point for when main watcharr/jellyfin btns break. */
		@media screen and (max-width: 320px) {
			flex-wrap: wrap;
		}

		button {
			display: flex;
			flex-flow: row;
			gap: 10px;
			text-transform: capitalize;

			.watcharr {
				font-family: "Rampart One";
				font-size: 19px;
				line-height: 19px;
			}

			&.other {
				overflow: hidden;
				animation: 250ms ease otherbtn;

				@keyframes otherbtn {
					from {
						width: 0px;
					}
					to {
						width: 100%;
					}
				}
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
