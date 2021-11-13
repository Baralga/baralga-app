<script>
  import {
    getProject,
    deleteProject,
    deleteProjectValidate
  } from "./stores.js";
  import { navigate } from "svelte-routing";

  export let id = undefined;
  let project;
  let projectTitle = "";
  let validation;

  function init() {
    getProject(id)
        .then((p) => {
          if (!p._links.delete) {
            back();
          }

          project = p;
          projectTitle = project.title;
        })
        .catch((err) => {
          return Promise.reject(err);
        });
    validation = deleteProjectValidate(project);
  }

  function back() {
    navigate("/projects", { replace: true });
  }

  function deleete() {
    deleteProject(project);
    back();
  }

  function cancel() {
    back();
  }

  init();
</script>

<style>

</style>

<h1 class="title is-1">Delete Project <span class="has-text-primary">{projectTitle}</span>?</h1>

<div class="columns is-multiline">

  <div class="column is-12">
    <p>Do you really want to delete the project?</p>
  </div>

  {#if validation.dependingActivitiesCount > 0}
    <div class="column is-12">
      <article class="message is-warning">
        <div class="message-body">
          Deleting the project will delete
          <strong>{validation.dependingActivitiesCount}</strong>
          activities you have tracked for that project.
        </div>
      </article>
    </div>
  {/if}

  <div class="column is-12">
    <div class="field is-grouped">
      <p class="control">

        <button class="button is-success" on:click={deleete}>Delete</button>
        <button class="button" on:click={cancel}>Cancel</button>

      </p>
    </div>
  </div>

</div>
