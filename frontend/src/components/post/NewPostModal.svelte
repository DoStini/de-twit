<script lang="ts">

    import {newPostModalStore} from "../../actions/newPostModal";
    import {serializeForm} from "../../utils/form";
    import type { FormValues} from "../../utils/form";

    export let submit: (formData: FormValues) => (Promise<boolean>);
    export let close: () => (void)
    export let loading: boolean

    let open: boolean;

    const onSubmit = async (evt) => {
        const formData = serializeForm(new FormData(evt.target));
        const success = await submit(formData);

        if (success) {
            evt.target.reset();
        }
    }

    const onClose = (evt) => {
        evt.target.parentNode.parentNode.reset();
        close()
    }

    newPostModalStore.subscribe((value) => open = value)
</script>

<input type="checkbox" bind:checked="{open}" id="post-modal" class="modal-toggle" />
<div class="modal">
    <div class="modal-box">
        <form id="new-post-form" on:submit|preventDefault={onSubmit}>
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
                <div on:click={onClose} class="btn text-error mr-2">Cancel</div>
                <button class="btn {loading ? 'loading' : ''}">Post!</button>
            </div>
        </form>
    </div>
</div>
