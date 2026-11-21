<script lang="ts">
	import { resolve } from "$app/paths";
	import Trans from "@/lib/i18n/Trans.svelte";
	import { t } from "@/lib/i18n/index.svelte";
</script>

<p>{t("handbook.jellyfinWebhook.intro")}</p>

<p>
	<Trans key="handbook.jellyfinWebhook.configIntro" />
</p>

<ol>
	<li>
		<p><Trans key="handbook.jellyfinWebhook.step1" /></p>
		<p><Trans key="handbook.jellyfinWebhook.step1Detail" /></p>
		<p><Trans key="handbook.jellyfinWebhook.step1Which" /></p>
	</li>
	<li>
		<p><Trans key="handbook.jellyfinWebhook.step2" /></p>
		<p>{t("handbook.jellyfinWebhook.step2Restart")}</p>
	</li>
	<li>
		<p>
			<Trans key="handbook.jellyfinWebhook.step3">
				{#snippet link(text: string)}
					<a href={resolve("/server")}>{text}</a>
				{/snippet}
			</Trans>
		</p>
	</li>
	<li>
		<p><Trans key="handbook.jellyfinWebhook.step4" /></p>
	</li>
	<li>
		<p><Trans key="handbook.jellyfinWebhook.step5" /></p>
	</li>
	<li>
		<p><Trans key="handbook.jellyfinWebhook.step6" /></p>
	</li>
	<li>
		<p><Trans key="handbook.jellyfinWebhook.step7" /></p>
		<p><Trans key="handbook.jellyfinWebhook.step7Note" /></p>
	</li>
	<li>
		<p><Trans key="handbook.jellyfinWebhook.step8" /></p>
		<ul>
			<li>
				<p>Playback Start</p>
			</li>

			<li>
				<p>Playback Stop</p>
			</li>

			<li>
				<p>User Data Saved</p>
			</li>
		</ul>
	</li>
	<li>
		<p><Trans key="handbook.jellyfinWebhook.step9" /></p>
	</li>
	<li>
		<p><Trans key="handbook.jellyfinWebhook.step10" /></p>
	</li>
	<li>
		<p><Trans key="handbook.jellyfinWebhook.step11" /></p>
	</li>
	<li>
		<p>{t("handbook.jellyfinWebhook.done")}</p>
	</li>
</ol>

<h2>{t("handbook.jellyfinWebhook.problems")}</h2>

<p>{t("handbook.jellyfinWebhook.problemsP1")}</p>

<p>{t("handbook.jellyfinWebhook.problemsP2")}</p>
