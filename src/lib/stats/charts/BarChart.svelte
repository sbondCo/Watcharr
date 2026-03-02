<script lang="ts">
	import { onMount } from "svelte";
	import {
		Chart,
		BarController,
		BarElement,
		CategoryScale,
		LinearScale,
		Tooltip,
	} from "chart.js";

	Chart.register(
		BarController,
		BarElement,
		CategoryScale,
		LinearScale,
		Tooltip,
	);

	interface Props {
		labels: string[];
		values: number[];
		colors?: string[];
	}

	let { labels, values, colors }: Props = $props();

	let canvas: HTMLCanvasElement;
	let chart: Chart | undefined;

	function getThemeColors() {
		const style = getComputedStyle(document.documentElement);
		return {
			text: style.getPropertyValue("--text-color").trim(),
			accent: style.getPropertyValue("--accent-color-hover").trim(),
			grid: style.getPropertyValue("--accent-color").trim(),
		};
	}

	function createChart() {
		if (chart) chart.destroy();
		const theme = getThemeColors();
		const barColors =
			colors && colors.length === labels.length
				? colors
				: labels.map(() => theme.accent);

		chart = new Chart(canvas, {
			type: "bar",
			data: {
				labels,
				datasets: [
					{
						data: values,
						backgroundColor: barColors,
						borderRadius: 4,
					},
				],
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				plugins: {
					legend: { display: false },
					tooltip: {
						callbacks: {
							label: (ctx) => `${ctx.parsed.y}`,
						},
					},
				},
				scales: {
					x: {
						ticks: { color: theme.text },
						grid: { display: false },
					},
					y: {
						beginAtZero: true,
						ticks: {
							color: theme.text,
							precision: 0,
						},
						grid: { color: theme.grid },
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
		// Track reactive dependencies
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
