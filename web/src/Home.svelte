<script>
  import { Router, Link, link, Route } from "svelte-routing";
  import moment from "moment/src/moment";
  import { projectStore, addActivity, reloadProjects } from "./stores.js";
  import { formatDuration } from "./formatter.js";
  import Projects from "./Projects.svelte";
  import Report from "./Report.svelte";
  import Activities from "./Activities.svelte";
  import ExportActivitiesAsCsv from "./ExportActivitiesAsCsv.svelte";
  import Home from "./Home.svelte";
  import { onMount } from "svelte";

  let runningActivity = false;
  let startTime;
  let activeProjectId;
  let selectedProject;
  let description;
  let durationFormatted;

  let updateTimerFunction;

  function updateTimer() {
    let duration = moment.duration(moment().diff(startTime));
    durationFormatted = formatDuration(duration) + " h";
    updateTimerFunction = setTimeout(updateTimer, 5000);
  }

  function stopActivity() {
    runningActivity = false;

    let activity = {
      startTime: startTime,
      endTime: moment().startOf("minute"),
      projectId: activeProjectId,
      description: description
    };

    addActivity(activity);

    durationFormatted = null;
    startTime = null;
    description = null;

    clearTimeout(updateTimerFunction);
  }

  function startActivity() {
    runningActivity = true;
    startTime = moment().startOf("minute");
    activeProjectId = selectedProject;
    updateTimer();
  }

  onMount(function() {
    reloadProjects();
  });
</script>

<style>
#x {
  padding: 0 1em 0 1em;
}
</style>

<div id="x" class="columns is-multiline">

  <div class="column is-12">
    <div class="buttons">
      <Router>
        <a
          href="/projects"
          use:link
          replace
          class="button is-link is-hidden-mobile">
          <span class="icon">
            <i class="fas fa-edit" />
          </span>
          <span>Manage Projects</span>
        </a>

        <a
          href="/projects"
          use:link
          replace
          class="button is-link is-hidden-desktop is-hidden-tablet">
          <span class="icon">
            <i class="fas fa-edit" />
          </span>
        </a>

        <a
          href="/activities/new"
          use:link
          replace
          class="button is-link is-hidden-mobile">
          <span class="icon">
            <i class="fas fa-plus" />
          </span>
          <span>Add Activity</span>
        </a>

        <a
          href="/activities/new"
          use:link
          replace
          class="button is-link is-hidden-desktop is-hidden-tablet">
          <span class="icon">
            <i class="fas fa-plus" />
          </span>
        </a>

      </Router>

      <ExportActivitiesAsCsv />
    </div>
  </div>

  <div class="column is-12 notification is-primary">
    <div class="columns is-multiline">
      <div class="column is-4">
        {#if runningActivity}
          <div on:click={stopActivity} class="button is-medium">
            <span class="icon is-medium">
              <i class="fa fa-stop" />
            </span>
          </div>
        {:else}
          <div on:click={startActivity} class="button is-large">
            <span class="icon is-medium">
              <i class="fa fa-play" />
            </span>
          </div>
        {/if}

        <div class="select is-large">
          <fieldset disabled={runningActivity}>
            <select bind:value={selectedProject}>
              {#each $projectStore as project}
                <option value={project.id}>{project.title}</option>
              {/each}
            </select>
          </fieldset>
        </div>

        <h3 class="title is-3">
          Start: {(startTime && startTime.format('HH:mm')) || '-'}
          <br />
          Duration: {durationFormatted || '-'}
        </h3>
      </div>
      <div class="column is-6">
        <fieldset disabled={!runningActivity}>
          <textarea
            bind:value={description}
            class="textarea is-primary"
            placeholder="Describe what you do ..." />
        </fieldset>
      </div>
    </div>
  </div>
</div>


<Report/>