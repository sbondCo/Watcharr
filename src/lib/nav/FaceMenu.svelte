<script lang="ts">
	import { store } from "@/store.svelte";
	import Menu from "../Menu.svelte";
	import { parseTokenPayload, userHasPermission } from "../util/helpers";
	import { UserPermission, UserType } from "@/types";
	import ProxyUserLogoutModal from "../logout/ProxyUserLogoutModal.svelte";
	import { goto } from "$app/navigation";
	import { clearWatcharrData } from "../logout";
	import { notify } from "../util/notify";
	import Modal from "../Modal.svelte";
	import Icon from "../Icon.svelte";

	let user = $derived(store.userInfo);
	let proxyUserLogoutShown = $state(false);
	let aboutModalOpen = $state(false);

	function logout() {
		if (user?.type === UserType.Proxy) {
			// Proxy users logout flow is different.
			proxyUserLogoutShown = true;
			return;
		}
		clearWatcharrData();
		goto("/login");
	}

	function profile() {
		goto("/profile");
	}

	function serverSettings() {
		goto("/server");
	}

	function userManagement() {
		goto("/manage_users");
	}

	function requestManagement() {
		goto("/arr_requests");
	}

	function shareWatchedList() {
		const nid = notify({ type: "loading", text: "Getting link" });
		const ud = parseTokenPayload();
		console.log(ud);
		if (ud?.userId && ud?.username) {
			const shareLink = `${window.location.origin}/lists/${ud.userId}/${ud.username}`;
			navigator.clipboard
				.writeText(shareLink)
				.then(() => {
					notify({ id: nid, type: "success", text: "Copied share link" });
				})
				.catch((r) => {
					console.error("Failed to copy list share link", r);
					notify({
						id: nid,
						type: "error",
						text: `Failed to copy share link:<br/><a href="${shareLink}" target="_blank">${shareLink}</a>`,
						time: 20000,
					});
				});
		} else {
			notify({ id: nid, type: "error", text: "Failed to get link" });
		}
	}

	function closeAbout() {
		aboutModalOpen = false;
	}
</script>

<Menu conf={{ arrowRight: "10px" }}>
	{#if user?.username}
		<h5 title={user.username}>Hi {user.username}!</h5>
	{/if}
	<button class="plain" onclick={() => profile()}>Profile</button>
	{#if !store.userSettings?.private}
		<button class="plain" onclick={() => shareWatchedList()}>Share List</button>
	{/if}
	{#if user && userHasPermission(user.permissions, UserPermission.PERM_ADMIN)}
		<button class="plain" onclick={() => serverSettings()}>Settings</button>
		<button class="plain" onclick={() => userManagement()}>Users</button>
		{#if store.serverFeatures?.sonarr || store.serverFeatures?.radarr}
			<!-- At least one (sonarr/radarr) should be enabled for requests menu item to display. -->
			<button class="plain" onclick={() => requestManagement()}>
				Requests
			</button>
		{/if}
	{/if}
	<button class="plain" onclick={() => logout()}>Logout</button>
	{#if proxyUserLogoutShown}
		<ProxyUserLogoutModal onClose={() => (proxyUserLogoutShown = false)} />
	{/if}
	<span style="color:black">
		<a
			class="menu-footer"
			href="#"
			onclick={() => {
				aboutModalOpen = !aboutModalOpen;
			}}
		>
			about
		</a>
		|
		<a
			class="menu-footer"
			href="https://github.com/sbondCo/Watcharr/releases"
			target="_blank"
		>
			v{__WATCHARR_VERSION__}
		</a>
	</span>
</Menu>

{#if aboutModalOpen}
	<Modal title="" onClose={closeAbout} maxWidth="400px">
		<div class="content">
			<div class="inner">
				<h2>About Watcharr</h2>
				<h5 class="norm">
					Watcharr is an open source project developed by <a
						href="https://github.com/sbondCo/Watcharr/commits?author=IRHM"
						>IRHM</a
					>
					and other great people.
				</h5>
				<div class="horizontal-icon-list">
					<a
						href="https://github.com/sbondCo/Watcharr"
						target="_blank"
						title="Github"
					>
						<Icon i="github" wh={60} />
					</a>
					<a
						href="https://watcharr.app/"
						target="_blank"
						title="Watcharr Project Page"
					>
						<Icon i="website" wh={60} />
					</a>
				</div>
				<h5 class="norm">Watcharr uses following media databases</h5>
				<div class="horizontal-icon-list">
					<a href="https://www.themoviedb.org/" target="_blank" title="TMDB">
						<Icon i="tmdb" wh={60} />
					</a>
					<a href="https://www.igdb.com/" target="_blank" title="IGDB">
						<Icon i="igdb" wh={60} />
					</a>
				</div>
			</div>
		</div>
	</Modal>
{/if}

<style lang="scss">
	h2.norm {
		margin-bottom: 16px;
	}

	h5.norm {
		margin: 8px 25px;
		text-align: center;
	}

	h5.norm > a {
		text-decoration: underline;
	}

	div.horizontal-icon-list {
		display: flex;
		justify-content: center;
		align-items: center;
		padding-bottom: 20px;
		padding-top: 5px;
	}

	div.horizontal-icon-list > a {
		display: inline;
		margin: 0 10px;
	}
</style>
