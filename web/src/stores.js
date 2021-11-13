import moment from "moment/src/moment";
import { writable, get } from "svelte/store";
import shortid from "./shortid";
import Api from "./api";

const createWritableStore = (key, startValue, elementInitializer) => {
  const { subscribe, set } = writable(startValue);

  if (elementInitializer) {
    const json = localStorage.getItem(key);
    if (json) {
      let storedObject = JSON.parse(json);
      set(storedObject);
      if (Array.isArray(storedObject) && elementInitializer) {
        storedObject.map(elementInitializer);
        set(storedObject);
      }
    }
  }

  return {
    subscribe,
    set,
    useLocalStorage: () => {
      const json = localStorage.getItem(key);
      if (json) {
        let storedObject = JSON.parse(json);
        set(storedObject);
        if (Array.isArray(storedObject) && elementInitializer) {
          storedObject.map(elementInitializer);
          set(storedObject);
        } else if (elementInitializer) {
          set(elementInitializer(storedObject));
        }
      }

      subscribe((current) => {
        localStorage.setItem(key, JSON.stringify(current));
      });
    },
  };
};

let projects = [
  {
    id: shortid(),
    name: "My Project",
  },
];

export const projectStore = createWritableStore("projects", projects);
projectStore.useLocalStorage();

const activityInitializer = (activity) => {
  if (!activity.id) {
    activity.id = shortid();
  }
  activity.startTime = moment(activity.startTime);
  activity.endTime = moment(activity.endTime);
  return activity;
};

export const activitiesStore = createWritableStore(
  "activities",
  [],
  activityInitializer
);
activitiesStore.useLocalStorage();

export const deleteProjectValidate = (project) => {
  let dependingActivities = [...get(activitiesStore)].filter(
    (activity) => activity.project.id === project.id
  );
  return {
    project: project,
    dependingActivities: dependingActivities,
    dependingActivitiesCount: dependingActivities.length,
  };
};

export const deleteProject = (project) => {
  Api.delete("/api/projects/" + project.id).then((data) => {
    reloadProjects();
  });
};

export const addActivity = (activity) => {
  let activityJson = {
    start: activity.startTime.format("yyyy-MM-DDTHH:mm:ss"),
    end: activity.endTime.format("yyyy-MM-DDTHH:mm:ss"),
    description: activity.description,
    _links: {
      project: {
        href: "/" + activity.projectId,
      }
    }
  }

  Api.post("/api/activities", activityJson).then((data) => {
    applyFilter(get(filterStore));
  });
};

export const updateActivity = (activity) => {
  let activityJson = {
    id: activity.id,
    start: activity.startTime.format("yyyy-MM-DDTHH:mm:ss"),
    end: activity.endTime.format("yyyy-MM-DDTHH:mm:ss"),
    description: activity.description,
    _links: {
      project: {
        href: "/" + activity.projectId,
      }
    }
  }

  Api.patch("/api/activities/" + activity.id, activityJson).then((data) => {
    applyFilter(get(filterStore));
  });
};

export const getActivity = (id) => {
  return Api.get("/api/activities/" + id);
};

const filterInitializer = (filter) => {
  filter.from = moment(filter.from);
  filter.to = moment(filter.to);
  return filter;
};

export const importBackup = (dataBackup) => {
  projectStore.set(dataBackup.projects);
  activitiesStore.set(dataBackup.activities);
};

export const filteredActivitiesStore = createWritableStore(
  "filteredActivities",
  []
);
let filter = {
  timespan: "year",
  from: moment().startOf("year"),
  to: moment().endOf("year"),
};
export const filterStore = createWritableStore(
  "filter",
  filter,
  filterInitializer
);
filterStore.useLocalStorage();

export const reloadProjects = () => {
  Api.get("/api/projects").then((data) => {
    projectStore.set(data._embedded.projects);
  });
};

export const applyFilter = (filter) => {
  var searchParams = new URLSearchParams();
  searchParams.set("start", filter.from.format("YYYY-MM-DD"))
  searchParams.set("end", filter.to.format("YYYY-MM-DD"))

  Api.get("/api/activities?" + searchParams.toString()).then((data) => {
    let embeddedActivities = data._embedded ? data._embedded.activities : [];
    let embeddedProjects = data._embedded ? data._embedded.projects : [];

    let projects = embeddedProjects.map((p) => {
      return {
        id: p.id,
        title: p.title,
        active: p.active,
        description: p.description,
      };
    });

    let activities = embeddedActivities.map((a) => {
      let projectId = a._links.project.href.substring(
        a._links.project.href.lastIndexOf("/") + 1
      );
      let project = projects.find((p) => p.id === projectId);

      return {
        id: a.id,
        description: a.description,
        startTime: moment(a.start),
        endTime: moment(a.end),
        project: project,
      };
    });

    filteredActivitiesStore.set(activities);
    filterStore.set(filter);
  });

};

export const addProject = (project) => {
  Api.post("/api/projects", project).then((data) => {
    reloadProjects();
  });
};

export const getProject = (id) => {
  return Api.get("/api/projects/" +  id);
};

export const totalDuration = () => {
  if (get(filteredActivitiesStore).length == 0) {
    return moment.duration(0);
  }

  let totalDuration = get(filteredActivitiesStore)
    .map((activity) =>
      moment.duration(activity.endTime.diff(activity.startTime))
    )
    .reduce((total, currentValue) => {
      return total.add(currentValue);
    });

  return totalDuration;
};

export const totalDurationStore = writable(totalDuration());

filteredActivitiesStore.subscribe(() => {
  totalDurationStore.set(totalDuration());
});
