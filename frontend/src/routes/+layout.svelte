<script lang="ts">
    import Navbar from "../components/Navbar.svelte";
    import SearchUserModal from "../components/users/SearchUserModal.svelte";

</script>

<div class="bg-base-200 absolute w-full h-full">
    <Navbar/>
    <div class="bg-base-200 absolute w-full">
<!--        <Notifications/>-->
        <slot></slot>
    </div>
</div>

<SearchUserModal/>
