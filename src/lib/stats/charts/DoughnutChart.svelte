<script lang="ts">
	import { onMount } from "svelte";
	import {
		Chart,
		DoughnutController,
		ArcElement,
		Tooltip,
		Legend,
	} from "chart.js";

	Chart.register(DoughnutController, ArcElement, Tooltip, Legend);

	interface Props {
		labels: string[];
		values: number[];
		colors: string[];
	}

	let { labels, values, colors }: Props = $props();

	let canvas: HTMLCanvasElement;
	let chart: Chart | undefined;

	function getThemeColors() {
		const style = getComputedStyle(document.documentElement);
		return {
			text: style.getPropertyValue("--text-color").trim(),
		};
	}

	function createChart() {
		if (chart) chart.destroy();
		const theme = getThemeColors();

		chart = new Chart(canvas, {
			type: "doughnut",
			data: {
				labels,
				datasets: [
					{
						data: values,
						backgroundColor: colors,
						borderWidth: 0,
					},
				],
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				cutout: "60%",
				plugins: {
					legend: {
						position: "right",
						labels: {
							color: theme.text,
							padding: 16,
							usePointStyle: true,
							pointStyleWidth: 12,
						},
					},
					tooltip: {
						callbacks: {
							label: (ctx) => ` ${ctx.label}: ${ctx.parsed}`,
						},
					},
				},
			},
		});
	}

	onMount(() => {
		createChart();
		return () => chart?.destroy();
	});

	$effect(() => {
		labels;
		values;
		colors;
		if (canvas) createChart();
	});
</script>

<div class="chart-container">
	<canvas bind:this={canvas}></canvas>
</div>

<style lang="scss">
	.chart-container {
		position: relative;
		width: 100%;
		height: 300px;
	}
</style>
