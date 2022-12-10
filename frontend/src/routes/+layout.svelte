<script>
    import Navbar from "../components/Navbar.svelte";
    import Notifications from "../components/alerts/Notifications.svelte";
    import NewPost from "../components/post/NewPostModal.svelte";
</script>

<div class="bg-base-200 absolute w-full h-full">
    <Navbar/>
    <div class="bg-base-200 absolute w-full">
<!--        <Notifications/>-->
        <slot></slot>
    </div>
</div>
