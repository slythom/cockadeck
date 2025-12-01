<script lang="ts">
	import { Sidebar, SidebarGroup, SidebarItem, SidebarButton, uiHelpers } from 'flowbite-svelte';
	import { ChartOutline, GridSolid, MailBoxSolid, UserSolid } from 'flowbite-svelte-icons';
	import { Badge, Button, Input, Label, Helper, Card, Toast } from 'flowbite-svelte';
	import {
		CheckCircleSolid,
		ExclamationCircleSolid,
		FireOutline,
		CloseCircleSolid
	} from 'flowbite-svelte-icons';
	import type { PageProps } from './$types';
	import { enhance } from '$app/forms';

	let { data, form }: PageProps = $props();

	let searchResults = $state<any[]>([]);
	import { page } from '$app/state';
	let activeUrl = $state(page.url.pathname);
	// import PlusPlaceholder from "$utils/PlusPlaceholder.svelte";
	const demoSidebarUi = uiHelpers();
	let isDemoOpen = $state(false);
	const closeDemoSidebar = demoSidebarUi.close;
	$effect(() => {
		isDemoOpen = demoSidebarUi.isOpen;
		activeUrl = page.url.pathname;
	});
	const spanClass = 'flex-1 ms-3 whitespace-nowrap';
	const activeClass =
		'flex items-center p-2 text-base font-normal text-white bg-primary-600 dark:bg-primary-700 rounded-lg dark:text-white hover:bg-primary-800 dark:hover:bg-primary-800';
	const nonActiveClass =
		'flex items-center p-2 text-base font-normal text-green-900 rounded-lg dark:text-white hover:bg-green-100 dark:hover:bg-green-700';
</script>

<SidebarButton onclick={demoSidebarUi.toggle} class="mb-2" />
<div class="flex h-screen">
	<Sidebar
		{activeUrl}
		backdrop={false}
		isOpen={isDemoOpen}
		closeSidebar={closeDemoSidebar}
		params={{ x: -50, duration: 50 }}
		classes={{ nonactive: nonActiveClass, active: activeClass }}
		class="w-64 flex-shrink-0"
	>
		<SidebarGroup>
			<SidebarItem label="Dashboard" href="/">
				{#snippet icon()}
					<ChartOutline
						class="h-5 w-5 text-gray-500 transition duration-75 group-hover:text-gray-900 dark:text-gray-400 dark:group-hover:text-white"
					/>
				{/snippet}
			</SidebarItem>
			<SidebarItem label="Kanban" {spanClass} href="/">
				{#snippet icon()}
					<GridSolid
						class="h-5 w-5 text-gray-500 transition duration-75 group-hover:text-gray-900 dark:text-gray-400 dark:group-hover:text-white"
					/>
				{/snippet}
				{#snippet subtext()}
					<span
						class="ms-3 inline-flex items-center justify-center rounded-full bg-gray-200 px-2 text-sm font-medium text-gray-800 dark:bg-gray-700 dark:text-gray-300"
						>Pro</span
					>
				{/snippet}
			</SidebarItem>
			<SidebarItem label="Inbox" {spanClass} href="/">
				{#snippet icon()}
					<MailBoxSolid
						class="h-5 w-5 text-gray-500 transition duration-75 group-hover:text-gray-900 dark:text-gray-400 dark:group-hover:text-white"
					/>
				{/snippet}
				{#snippet subtext()}
					<span
						class="ms-3 inline-flex h-3 w-3 items-center justify-center rounded-full bg-primary-200 p-3 text-sm font-medium text-primary-600 dark:bg-primary-900 dark:text-primary-200"
						>3</span
					>
				{/snippet}
			</SidebarItem>
			<SidebarItem label="Sidebar" href="/docs/components/sidebar">
				{#snippet icon()}
					<UserSolid class="h-5 w-5 text-primary-500 transition duration-75 " />
				{/snippet}
			</SidebarItem>
		</SidebarGroup>
	</Sidebar>
	<div class="flex-1 overflow-auto px-4">
		<div class="rounded-lg border-2 border-dashed border-gray-200 p-4 dark:border-gray-700">
			<div class="flex gap-4 p-6">
				<Card class="max-w-none flex-1 p-4">
					<h2 class="mb-4 text-xl font-bold text-slate-300">Create a deck/collection</h2>
					<form action="?/createCollection" method="POST" use:enhance>
						<div class="mb-6 grid grid-cols-[1fr_auto] gap-4">
							<div>
								<Label for="name" class="mb-2">Choose a name:</Label>
								<Input id="name" type="text" name="name" required />
							</div>
							<div>
								<Button type="submit" class="mt-7 cursor-pointer">Create</Button>
							</div>
						</div>
					</form>
				</Card>

				<Card class="max-w-none flex-1 p-4">
					<h2 class="mb-4 text-xl font-bold text-slate-300">Search your cards</h2>
					<div class="mx-auto">
						<form
							action="?/search"
							method="POST"
							class="mx-auto"
							use:enhance={() => {
								return async ({ result, update }) => {
									await update();
									if (
										result.type === 'success' &&
										result.data?.context === 'search' &&
										result.data?.card_name
									) {
										searchResults = [
											{
												name: result.data.card_name,
												setcode: result.data.card_setcode,
												collector_number: result.data.card_collector_number,
												set_name: result.data.card_set_name,
												image: result.data.card_image_normal,
												colors: result.data.card_colors,
												mana_cost: result.data.card_mana_cost,
												quantity: result.data.card_quantity
											},
											...searchResults
										];
									}
								};
							}}
						>
							<div class="mb-6 grid grid-cols-[1fr_1fr_1fr_auto] gap-4">
								<div>
									<Label for="setcode" class="mb-2">Set Code</Label>
									<Input id="setcode" type="text" name="setcode" required />
								</div>

								<div>
									<Label for="collector_number" class="mb-2">Collector Number</Label>
									<Input id="collector_number" type="text" name="collector_number" required />
								</div>
								<div>
									<Label for="quantity" class="mb-2">Quantity</Label>
									<Input id="quantity" type="number" name="quantity" required />
								</div>
								<div>
									<Button type="submit" class="mt-7 cursor-pointer">Find!</Button>
								</div>
							</div>
						</form>
					</div>
				</Card>
			</div>
		</div>
	</div>
</div>
