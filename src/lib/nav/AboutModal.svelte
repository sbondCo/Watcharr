<script lang="ts">
	import Icon from "../Icon.svelte";
	import Modal from "../Modal.svelte";

	interface Props {
		onClose: () => void;
	}

	let { onClose }: Props = $props();
</script>

<Modal title="" {onClose} maxWidth="400px">
	<div class="content">
		<div class="inner">
			<h2>About Watcharr</h2>
			<h5 class="norm">
				Watcharr is an open source project developed by <a
					href="https://github.com/sbondCo/Watcharr/commits?author=IRHM">IRHM</a
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

<style lang="scss">
	h2 {
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
