<script lang="ts">
	import { Navbar, NavBrand, NavLi, NavUl, NavHamburger, Button } from 'flowbite-svelte';
</script>

<Navbar>
	<NavBrand href="/">
		<!-- <img src="/images/flowbite-svelte-icon-logo.svg" class="me-3 h-6 sm:h-9" alt="logo" /> -->
		<span class="self-center text-xl font-semibold whitespace-nowrap dark:text-white"
			>Cockadeck</span
		>
	</NavBrand>
	<NavUl class="order-1">
		<NavLi href="/">Home</NavLi>
		<NavLi href="/search">Search</NavLi>
		<NavLi href="/account/my-cards">My cards</NavLi>
		<NavLi href="/account/connect">Account</NavLi>
	</NavUl>
</Navbar>
