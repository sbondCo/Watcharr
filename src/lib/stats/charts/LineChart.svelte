<script lang="ts">
	import { onMount } from "svelte";
	import {
		Chart,
		LineController,
		LineElement,
		PointElement,
		CategoryScale,
		LinearScale,
		Tooltip,
		Filler,
	} from "chart.js";

	Chart.register(
		LineController,
		LineElement,
		PointElement,
		CategoryScale,
		LinearScale,
		Tooltip,
		Filler,
	);

	interface Props {
		labels: string[];
		values: number[];
	}

	let { labels, values }: Props = $props();

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

		chart = new Chart(canvas, {
			type: "line",
			data: {
				labels,
				datasets: [
					{
						data: values,
						borderColor: theme.accent,
						backgroundColor: theme.accent + "33",
						fill: true,
						tension: 0.3,
						pointRadius: 3,
						pointHoverRadius: 6,
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
						ticks: { color: theme.text, maxRotation: 45 },
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
		labels;
		values;
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
