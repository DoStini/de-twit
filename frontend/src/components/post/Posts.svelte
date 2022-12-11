<script lang="ts">
    import Post from "./Post.svelte";
    import type PostData from "../../types/PostData.js";
    import NewPostsBadge from "./NewPostsBadge.svelte";
    import {env} from "$env/dynamic/public";

    export let posts: [PostData];
    export let newPosts: [PostData];
    export let mainTimeline: boolean
    export let emptyPostsMessage: string
    export let refreshTimeline : () => (void)

    const refreshPosts = () => {
        refreshTimeline()
        window.scrollTo({top: 0, behavior: 'smooth'});
    }

</script>

{ #if newPosts.length > 0}
    <NewPostsBadge action={refreshPosts} newPostsCount={newPosts.length} />
{/if}

{ #if posts.length === 0 }
    <div class="flex items-center justify-center m-[5em]">
        <span class="text-xl">
            { emptyPostsMessage }
        </span>
    </div>
{:else }
    <div class={`posts-card-grid${mainTimeline ? '' : '-center'}`}>

    {#each posts as post}
        {@const isUser = post.username === env.PUBLIC_USERNAME }
        {@const classname = mainTimeline ? (isUser ? 'card-show-user' : 'card-show-other') : 'card-show-center' }
        <div class="{classname}">
            <Post post="{post}"></Post>
        </div>

    {/each}
    </div>
{/if}
