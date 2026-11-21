<script lang="ts">
	import step1 from "./step1.png";
	import step2 from "./step2.png";
	import step3 from "./step3.png";
	import step4 from "./step4.png";
	import fullPage from "./content-full-page.png";
	import HandbookImage from "@/lib/handbook/internal/HandbookImage.svelte";
	import { t } from "@/lib/i18n/index.svelte";
</script>

<p>{t("handbook.addingToWatchlist.p1")}</p>

<p>{t("handbook.addingToWatchlist.p2")}</p>

<p>{t("handbook.addingToWatchlist.p3")}</p>

<ol>
	<li>
		<p>{t("handbook.addingToWatchlist.step1")}</p>
		<p>
			<HandbookImage src={step1} alt={t("handbook.step", { step: 1 })} />
		</p>
	</li>
	<li>
		<p>{t("handbook.addingToWatchlist.step2")}</p>
		<p>
			<HandbookImage src={step2} alt={t("handbook.step", { step: 2 })} />
		</p>
	</li>
	<li>
		<p>{t("handbook.addingToWatchlist.step3")}</p>
		<p>{t("handbook.addingToWatchlist.step3Statuses")}</p>
		<p>
			<HandbookImage src={step3} alt={t("handbook.step", { step: 3 })} />
		</p>
	</li>
	<li>
		<p>{t("handbook.addingToWatchlist.step4")}</p>
		<p>
			<HandbookImage src={step4} alt={t("handbook.step", { step: 4 })} />
		</p>
	</li>
</ol>

<p>{t("handbook.addingToWatchlist.p4")}</p>

<p>
	<HandbookImage
		src={fullPage}
		alt={t("handbook.addingToWatchlist.fullPageAlt")}
	/>
</p>

<p>{t("handbook.addingToWatchlist.p5")}</p>
