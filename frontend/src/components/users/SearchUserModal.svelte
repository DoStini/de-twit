<script lang="ts">

    import {serializeForm} from "../../utils/form";
    import type { FormValues} from "../../utils/form";
    import {closeSearchUserModal, searchUserModal} from "../../actions/searchUserModal";
    import type UserData from "../../types/UserData";
    import UserBadge from "./UserBadge.svelte";
    import {followUser, searchUser} from "../../services/users";
    import { clickOutside } from '../../utils/clickOutisde';

    let submit: (formData: FormValues) => (Promise<boolean>);
    let close: () => (void)
    let loading: boolean
    let user: UserData | null
    let error: boolean = false

    const resetUser = () => {
        user = null;
        error = false
    }

    const submitSearch = async (data) => {
        loading = true
        try {
            const userData = await searchUser(data.username);
            user = userData;
            loading = false;
            error = false
            return true;
        } catch (e) {
            console.error(e)
            loading = false;
            user = null;
            error = true
            return true;
        }
    }

    const submitFollow = async () => {
        loading = true
        try {
            const target = !user.following
            await followUser(user.username, target);
            user = {...user, following: target}
            loading = false;
            return true;
        } catch (e) {
            console.error(e)
            loading = false;
            return false;
        }
    }

    const handleClose = () => {
        closeSearchUserModal()
        resetUser()
    }

    let open: boolean;

    const resetForm = (evt: Event = null) => {
        const form: any = evt ? evt.target : document.getElementById("find-user-modal")

        setTimeout(form.reset.bind(form), 200)
    }

    const onSubmitSearch = async (evt) => {
        const formData = serializeForm(new FormData(evt.target));
        const success = await submitSearch(formData);

        if (success) {
            resetForm(evt)
        }
    }

    const onClose = () => {
        resetForm()
        handleClose()
    }

     searchUserModal.subscribe((value) => open = value)
</script>

<input type="checkbox" bind:checked="{open}" id="user-modal" class="modal-toggle" />
<div class="modal">
    <div class="modal-box" use:clickOutside={() => open && onClose()}>
        <form id="find-user-modal" on:submit|preventDefault={onSubmitSearch}>
            <h3 class="font-bold text-lg">Find friends!</h3>
            {#if user || error}
                <UserBadge bind:user={user} loading={loading} error={error} follow={submitFollow} />
                <div class="modal-action">
                    <div on:click={onClose} class="btn text-error mr-2">Cancel</div>
                    <div on:click={resetUser} class="btn {loading ? 'loading' : ''}">Search</div>
                </div>
            {:else}
                <input
                    id="find-username"
                    name="username"
                    form="find-user-modal"
                    class="textarea mt-5 textarea-bordered w-full"
                    placeholder="Username"
                    required
                >
                <div class="modal-action">
                    <div on:click={onClose} class="btn text-error mr-2">Cancel</div>
                    <button class="btn {loading ? 'loading' : ''}">Search</button>
                </div>
            {/if}
        </form>
    </div>
</div>
