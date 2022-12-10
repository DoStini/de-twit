
<script lang="ts">
    import type PostData from "../../types/PostData";

    export let post: PostData;
</script>

<div class={`card bg-base-100 shadow-xl`}>
    <div class="card-body">
        <h2 class="card-title">{post.username}</h2>
        <p>{post.text}</p>
        <p>{post.timestamp?.toDateString()}</p>
    </div>
</div>
