<script lang="ts">
	import { t } from "@/lib/i18n/index.svelte";
	import Icon from "@/lib/Icon.svelte";
	import Modal from "@/lib/Modal.svelte";
	import Trans from "@/lib/i18n/Trans.svelte";

	let modalShown = $state(false);
</script>

<button
	onclick={() => (modalShown = true)}
	style="width: min-content; margin-left:auto; gap: 5px;"
>
	<Icon i="filter" wh={16} />
	{t("search.filters.title")}
</button>

{#if modalShown}
	<Modal title={t("search.filters.help")} onClose={() => (modalShown = false)}>
		<div class="modal-inner">
			<p>
				{t("search.filters.howTo")}
			</p>
			<p><b>{t("search.filters.supported")}</b></p>
			<ul>
				<li>
					<b>year</b>, <b>y</b> - <Trans key="search.filters.year" />
				</li>
				<li>
					<b>fyear</b>, <b>fy</b> - <Trans key="search.filters.fyear" />
				</li>
				<li>
					<b>adult</b> - {t("search.filters.adult")}
				</li>
			</ul>
			<p><b>{t("search.filters.examples")}</b></p>
			<ul>
				<li>
					<b>{t("search.filters.exampleYear")}</b>
					<ul>
						<li>Spider Man year:2026</li>
						<li>Spider Man y:2026</li>
					</ul>
				</li>
				<li>
					<b>{t("search.filters.exampleFyear")}</b>
					<ul>
						<li>Spider Man fyear:2026</li>
						<li>Spider Man fy:2026</li>
					</ul>
				</li>
				<li>
					<b>{t("search.filters.exampleAdult")}</b>
					<ul>
						<li>Spider Man adult:1</li>
					</ul>
				</li>
			</ul>
		</div>
	</Modal>
{/if}

<style lang="scss">
	.modal-inner {
		display: flex;
		flex-flow: column;
		gap: 5px;

		ul {
			list-style: inside;
			margin-left: 5px;
		}

		ul li > ul {
			list-style: inside circle;
			margin-left: 15px;
		}
	}
</style>
