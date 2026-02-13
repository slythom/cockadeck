<script lang="ts">
	import { enhance } from '$app/forms';
	import { FloatingLabelInput, Helper, Button } from 'flowbite-svelte';
</script>

<div class="mx-auto">
	<div id="search-card-form" class="bg-red-500 space-y-4 p-10">
		<form action="?/search" method="POST" use:enhance>
		<FloatingLabelInput variant="filled" id="set_code" name="set_code" type="text"
			>set_code</FloatingLabelInput
		>
    <FloatingLabelInput variant="filled" id="collector_number" name="collector_number" type="text"
			>collector_number</FloatingLabelInput
		>
		<Button type="submit">Search</Button>
		</form>
	</div>
</div>

