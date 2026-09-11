<script lang="ts">
	import step1 from "./step1.png";
	import step2 from "./step2.png";
	import step3 from "./step3.png";
	import step4 from "./step4.png";
	import fullPage from "./content-full-page.png";
	import HandbookImage from "@/lib/handbook/internal/HandbookImage.svelte";
</script>

<p>Adding media to your list is the core feature of Watcharr.</p>

<p>
	The homepage is where your list will appear once you have added something.
</p>

<p>
	To get started, you can follow along with these steps to add your first movie
	(or show, etc):
</p>

<ol>
	<li>
		<p>Use the search bar at the top of the page to find content.</p>
		<p>
			<HandbookImage src={step1} alt="Step 1" />
		</p>
	</li>
	<li>
		<p>Hover or click on the content you want to add.</p>
		<p>
			<HandbookImage src={step2} alt="Step 2" />
		</p>
	</li>
	<li>
		<p>Clicking the + (Plus) icon will let you give the content a status.</p>
		<p>
			Currently supported statuses include: Finished, Watching, Held, Planned,
			Dropped.
		</p>
		<p>
			<HandbookImage src={step3} alt="Step 3" />
		</p>
	</li>
	<li>
		<p>Rate the content.</p>
		<p>
			<HandbookImage src={step4} alt="Step 4" />
		</p>
	</li>
</ol>

<p>
	If you prefer, you can also do this by clicking/tapping on the active poster
	again. This will take you to the contents full page. Here you can see all the
	options, including adding Thoughts for the content.
</p>

<p>
	<HandbookImage src={fullPage} alt="Content Full Page" />
</p>

<p>
	Thoughts can be anything you would like to write down or remember about the
	specific movie/show. For example, sometimes I like to write something specific
	that I really liked about the content.
</p>
