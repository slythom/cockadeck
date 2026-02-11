<script lang="ts">
  import { enhance } from '$app/forms';
  import type { PageProps } from './$types';
  let { form, data }: PageProps = $props();
  import { Card, Button, Label, Input, Checkbox } from "flowbite-svelte";
</script>

<Card class="p-4 sm:p-6 md:p-8">
  <form class="flex flex-col space-y-6" action="?/signin" method="POST" use:enhance>
    <h3 class="text-xl font-medium text-gray-900 dark:text-white">Sign in to our platform</h3>
    <Label class="space-y-2">
      <span>Your email</span>
      <Input type="email" name="email" placeholder="name@company.com" required />
    </Label>
    <Label class="space-y-2">
      <span>Your password</span>
      <Input type="password" name="password" placeholder="•••••" required />
    </Label>
    <div class="flex items-start">
      <Checkbox>Remember me</Checkbox>
      <a href="/" class="text-primary-700 dark:text-primary-500 ms-auto text-sm hover:underline">Lost password?</a>
    </div>
    <Button type="submit" class="w-full">Log in!</Button>
  </form>
</Card>
{#if form?.success === false}
  <p class="text-red-500 mt-4">{form.message}</p>
{/if}
{#if form?.success === true}
  <p class="text-green-500 mt-4">{form.message}. Welcome, {data.userEmail}.</p>
{/if}
