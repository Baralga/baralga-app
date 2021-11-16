<script>
  import moment from "moment/src/moment";
  import {
    projectStore,
    addActivity,
    updateActivity,
    getActivity,
    reloadProjects
  } from "./stores.js";
  import { navigate } from "svelte-routing";
  import { onMount } from "svelte";

  export let id = undefined;
  let projectId;
  let dateValue;
  let timeFromValue;
  let timeToValue;
  let description;

  function validateDate(dateValue) {
    let parsedDate = moment(dateValue, "DD.MM.YYYY", "de", true);
    return parsedDate !== undefined && parsedDate.isValid();
  }

  function validateTime(timeValue) {
    let parsedTime = moment(timeValue, "HH:mm", "de", true);
    return parsedTime !== undefined && parsedTime.isValid();
  }

  function validateAfter(from, to) {
    let parsedFrom = moment(from, "HH:mm", "de", true);
    let parsedTo = moment(to, "HH:mm", "de", true);
    return parsedFrom.isSameOrBefore(parsedTo);
  }

  $: isValidDateValue = validateDate(dateValue);
  $: isValidTimeFrom =
    validateTime(timeFromValue) && validateAfter(timeFromValue, timeToValue);
  $: isValidTimeTo =
    validateTime(timeToValue) && validateAfter(timeFromValue, timeToValue);

  function back() {
    navigate("/", { replace: true });
  }

  function isEditMode() {
    return id !== undefined;
  }

  function save() {
    let startTime = moment(dateValue, "DD.MM.YYYY", "de", true);
    let endTime = moment(dateValue, "DD.MM.YYYY", "de", true);

    let parsedTimeFrom = moment(timeFromValue, "HH:mm", "de", true);
    let parsedTimeTo = moment(timeToValue, "HH:mm", "de", true);

    startTime.set("hour", parsedTimeFrom.hours());
    startTime.set("minute", parsedTimeFrom.minutes());

    endTime.set("hour", parsedTimeTo.hours());
    endTime.set("minute", parsedTimeTo.minutes());

    let activity = {
      startTime: startTime,
      endTime: endTime,
      projectId: projectId,
      description: description
    };

    if (isEditMode()) {
      activity.id = id;
      updateActivity(activity).then(() => {
        back();
      });
    } else {
      addActivity(activity).then(() => {
        back();
      });
    }
  }

  onMount(function() {
    reloadProjects();
    init();
  });

  function init() {
    projectId = $projectStore[0].id;
    description = null;
    dateValue = moment()
      .startOf("minute")
      .format("DD.MM.YYYY");
    timeFromValue = moment().format("HH:mm");
    timeToValue = moment().format("HH:mm");

    if (id !== undefined) {
      getActivity(id)
        .then(activity => {
          let startTime = moment(activity.start);
          let endTime = moment(activity.end);

          let activityProjectId = activity._links.project.href.substring(
            activity._links.project.href.lastIndexOf("/") + 1
          );
          projectId = activityProjectId;
          description = activity.description;
          dateValue = startTime.startOf("minute").format("DD.MM.YYYY");
          timeFromValue = startTime.format("HH:mm");
          timeToValue = endTime.format("HH:mm");
          return;
        })
        .catch(error => {
          cancel();
        });
    }
  }

  function completeTimeFromValue(event) {
    timeFromValue = completeTimeValue(timeFromValue)
  }

  function completeTimeToValue(event) {
    timeToValue = completeTimeValue(timeToValue)
  }

  function completeTimeValue(time) {
    let completedTime = time;

    completedTime = completedTime.replace(",,", ":");
    completedTime = completedTime.replace('/', ':');
    completedTime = completedTime.replace(';', ',');
    completedTime = completedTime.replace('.', ':');

    // Treat 11,25 as 11:15
    // Treat 11,75 as 11:45
    // Treat 11,5 and 11,50 as 11:30
    let splittedTime = completedTime.split(",");
    if (completedTime.includes(",") && splittedTime.length >= 2) {
    	let hh = splittedTime[0];
    	let mm = splittedTime[1];
		if (mm.length < 2) {
			mm = mm + "0"; 
		}

  	try {
    		// Convert to integer value
    		let m = parseInt(mm);
    		// Convert to float for calculation
    		let fm = m;          
    		// Convert from base100 to base60
    		fm *= 0.6;                     
    		// Round to int
    		m = Math.round(fm);  
    		mm = "" + m;
    		if (mm.length < 2) {
    			mm = "0" + mm; 
    		}
    		if (hh.length < 2) {
    			hh = "0" + hh; 
    		}
    		completedTime = hh + ":" + mm;
    	} catch (e) {
    		// Conversion to int failed so smart format does not apply.
    	}
    }

    if (!completedTime.includes(":")) {
      completedTime += ":00"
    }

    return completedTime;
  }

  function cancel() {
    back();
  }

</script>

<style>

</style>

<h1 class="title is-1">
  {#if isEditMode()}Edit{:else}Add{/if}
  Activity
</h1>

<div class="field">
  <label class="label">Project</label>
  <div class="control select">
    <select bind:value={projectId}>
      {#each $projectStore as project}
        <option value={project.id}>{project.title}</option>
      {/each}
    </select>
  </div>
</div>

<div class="field">
  <label class="label">Date</label>
  <div class="control">
    <input
      pattern="[0-3][0-9]\.[0-1][0-9]\.20[0-9]{2}"
      minlength="10"
      maxlength="10"
      class="input"
      class:is-danger={!isValidDateValue}
      bind:value={dateValue}
      type="text"
      placeholder="16.11.2019" />
  </div>
</div>

<div class="field">
  <label class="label">Start Time</label>
  <div class="control">
    <input
      class="input"
      pattern="[0-9]{2}:[0-5][0-9]"
      bind:value={timeFromValue}
      on:change="{completeTimeFromValue}"
      class:is-danger={!isValidTimeFrom}
      minlength="5"
      maxlength="5"
      type="text"
      placeholder="10:00" />
  </div>
</div>

<div class="field">
  <label class="label">End Time</label>
  <div class="contol">
    <input
      class="input"
      pattern="[0-9]{2}:[0-5][0-9]"
      bind:value={timeToValue}
      on:change="{completeTimeToValue}"
      class:is-danger={!isValidTimeTo}
      minlength="5"
      maxlength="5"
      type="text"
      placeholder="10:00" />
  </div>
</div>

<div class="field">
  <label class="label">Description</label>
  <div class="control">
    <textarea
      bind:value={description}
      class="textarea"
      placeholder="Describe what you do ..." />
  </div>
</div>

<div class="field is-grouped">
  <p class="control">

    <button
      class="button is-success"
      disabled={!(isValidDateValue && isValidTimeFrom && isValidTimeTo)}
      on:click={save}>
      {#if isEditMode()}Update{:else}Add{/if}
    </button>
    <button class="button" on:click={cancel}>Cancel</button>

  </p>
</div>
