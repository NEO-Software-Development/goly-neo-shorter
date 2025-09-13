<script>
	/**
	 * @component
	 * The main application component.
	 * It serves as the root of the application and renders the main layout.
	 */
	import List from "./components/List.svelte"
	import New from "./components/New.svelte"
</script>

<main>
	<h1>Goly -- The Go URL Shortener!</h1>
	<List />
</main>
<New />
