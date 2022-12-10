
<script lang="ts">
    import type PostData from "../../types/PostData";
    import {dateToString} from "../../utils/date.js";

    export let post: PostData;
</script>

<div class={`card bg-base-100 shadow-xl`}>
    <div class="card-body">
        <h2 class="card-title">{post.username}</h2>
        <p>{post.text}</p>
        <p>{dateToString(post.timestamp)}</p>
    </div>
</div>
