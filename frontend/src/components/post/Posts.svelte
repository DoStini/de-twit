<script lang="ts">
    import Post from "./Post.svelte";
    import type PostData from "../../types/PostData.js";
    import NewPostsBadge from "./NewPostsBadge.svelte";
    import {refreshTimeline} from "../../actions/posts.js";

    export let posts: [PostData];
    export let newPosts: [PostData];

    const refreshPosts = () => {
        refreshTimeline()
        window.scrollTo({top: 0, behavior: 'smooth'});
    }

</script>

{ #if newPosts.length > 0}
    <NewPostsBadge action={refreshPosts} newPostsCount={newPosts.length} />
{/if}

{ #if posts.length === 0 }
    <div class="flex items-center justify-center m-5">
        <span class="text-xl">
            Nothing to show! Post something or find other users to follow!
        </span>
    </div>
{:else }
    <div class="grid grid-cols-8 gap-4 sm:m-20 m-5">

    {#each posts as post}
        {@const isUser = post.username === "andremoreira9" }
        <div class="{isUser ? 'card-show-user' : 'card-show-other'}">
            <Post post="{post}"></Post>
        </div>

    {/each}
    </div>
{/if}
