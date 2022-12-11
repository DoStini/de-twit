<script lang="ts">
    export let newPostsCount: number;
    export let action;
</script>

<div class="sticky top-5 flex items-center justify-center pt-5 z-10">
    <button class="btn btn-primary gap-2" on:click={action}>
        New Posts
        <div class="badge badge-accent">{newPostsCount > 99 ? "+99" : newPostsCount}</div>
    </button>
</div>
