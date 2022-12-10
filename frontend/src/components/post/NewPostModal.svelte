<script lang="ts">

    import {newPostModalStore} from "../../actions/newPostModal";
    import { serializeForm} from "../../utils/form";
    import type { FormValues} from "../../utils/form";

    export let submit: (formData: FormValues) => (void);

    let open: boolean;

    const onSubmit = (evt) => {
        const formData = serializeForm(new FormData(evt.target));
        submit(formData);

        evt.target.reset();
    }

    newPostModalStore.subscribe((value) => open = value)
    console.log(open)
</script>

<input type="checkbox" bind:checked="{open}" id="post-modal" class="modal-toggle" />
<div class="modal">
    <div class="modal-box">
        <form id="new-post-form" on:submit={onSubmit}>
            <h3 class="font-bold text-lg">Post something to your followers</h3>
                <textarea
                    id="post-content"
                    name="content"
                    form="new-post-form"
                    class="textarea mt-5 textarea-bordered w-full"
                    placeholder="What are you thinking about?"
                    rows="5"
                    style="resize: none"
                    required
                ></textarea>
            <div class="modal-action">
                <button type="submit" class="btn">Yay!</button>
            </div>
        </form>
    </div>
</div>
