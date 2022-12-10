<script lang="ts">
    import Post from "./Post.svelte";
    import type PostData from "../../types/PostData.js";

    export let posts: [PostData];
</script>

{ #if posts.length === 0 }
    <div class="flex items-center justify-center m-5">
        <span class="text-xl">
            Nothing to show! Post something or find other users to follow!
        </span>
    </div>
{:else }
    <div class="grid grid-cols-8 gap-4 sm:m-20 m-5">

    {#each posts as post}
        {@const colStartPosition = post.username === "andremoreira9" ? 2 : 3}
        <div class="sm:col-start-{colStartPosition} sm:col-span-5 col-span-12">
            <Post post="{post}"></Post>
        </div>

    {/each}
    </div>
{/if}
