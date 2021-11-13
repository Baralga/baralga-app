import moment from "moment/src/moment";

export const readXml = (xmlBackupContents) => {
  try {
    let parser = new DOMParser();
    let xmlDoc = parser.parseFromString(xmlBackupContents, "text/xml");

    let errElemCount = xmlDoc.getElementsByTagName("parsererror").length;
    if (errElemCount > 0) {
      throw xmlDoc.getElementsByTagName("parsererror");
    }

    let projects = [];
    let projectElements = xmlDoc.getElementsByTagName("project");
    for (let i = 0; i < projectElements.length; i++) {
      let projectElement = projectElements[i];

      let project = {
        id: projectElement.attributes.id.nodeValue,
        name: projectElement.getElementsByTagName("title")[0].innerHTML,
        description: projectElement.getElementsByTagName("description")[0]
          .innerHTML,
      };

      projects.push(project);
    }

    let activities = [];
    let activityElements = xmlDoc.getElementsByTagName("activity");
    for (let i = 0; i < activityElements.length; i++) {
      let activityElement = activityElements[i];

      let startTime = moment(
        activityElement.attributes.start.nodeValue,
        "yyyy-MM-DDTHH:mm"
      );
      let endTime = moment(
        activityElement.attributes.end.nodeValue,
        "yyyy-MM-DDTHH:mm"
      );

      let projectId = activityElement.attributes.projectReference.nodeValue;
      let project = projects.find((project) => project.id == projectId);

      let activity = {
        id: activityElement.attributes.id.nodeValue,
        startTime: startTime,
        endTime: endTime,
        project: project,
      };

      activities.push(activity);
    }

    return {
      status: "ok",
      activities: activities,
      projects: projects,
    };
  } catch (e) {
    return {
      status: "error",
      activities: [],
      projects: [],
    };
  }
};

export const createXml = (activities, projects) => {
  let xmlDoc = document.implementation.createDocument(null, "baralga");
  let baralgaElement = xmlDoc.getElementsByTagName("baralga")[0];

  let versionAttribute = xmlDoc.createAttribute("version");
  versionAttribute.value = "1";
  baralgaElement.setAttributeNode(versionAttribute);

  let projectsElement = xmlDoc.createElement("projects");
  baralgaElement.appendChild(projectsElement);

  projects.map((project) => {
    let projectElement = xmlDoc.createElement("project");
    projectsElement.appendChild(projectElement);

    let activeAttribute = xmlDoc.createAttribute("active");
    activeAttribute.value = "true";
    projectElement.setAttributeNode(activeAttribute);

    let idAttribute = xmlDoc.createAttribute("id");
    idAttribute.value = project.id;
    projectElement.setAttributeNode(idAttribute);

    let titleElement = xmlDoc.createElement("title");
    let titleText = xmlDoc.createTextNode(project.name);
    titleElement.appendChild(titleText);
    projectElement.appendChild(titleElement);

    let descriptionElement = xmlDoc.createElement("description");
    projectElement.appendChild(descriptionElement);

    if (project.description) {
      let descriptionText = xmlDoc.createTextNode(project.description);
      descriptionElement.appendChild(descriptionText);
    }
  });

  let activitiesElement = xmlDoc.createElement("activities");
  baralgaElement.appendChild(activitiesElement);

  activities.map((activity) => {
    let activityElement = xmlDoc.createElement("activity");
    activitiesElement.appendChild(activityElement);

    let idAttribute = xmlDoc.createAttribute("id");
    idAttribute.value = activity.id;
    activityElement.setAttributeNode(idAttribute);

    let projectReferenceAttribute = xmlDoc.createAttribute("projectReference");
    projectReferenceAttribute.value = activity.project.id;
    activityElement.setAttributeNode(projectReferenceAttribute);

    let startAttribute = xmlDoc.createAttribute("start");
    startAttribute.value = activity.startTime.format("yyyy-MM-DDTHH:mm");
    activityElement.setAttributeNode(startAttribute);

    let endAttribute = xmlDoc.createAttribute("end");
    endAttribute.value = activity.endTime.format("yyyy-MM-DDTHH:mm");
    activityElement.setAttributeNode(endAttribute);

    let descriptionElement = xmlDoc.createElement("description");
    if (activity.description) {
      let descriptionText = xmlDoc.createTextNode(activity.description);
      descriptionElement.appendChild(descriptionText);
    }
    activityElement.appendChild(descriptionElement);
  });

  let xmlString = new XMLSerializer().serializeToString(xmlDoc);

  let blob = new Blob(
    ['<?xml version="1.0" encoding="UTF-8" standalone="no"?>', xmlString],
    {
      type: "text/xml;charset=utf-8",
    }
  );

  return blob;
};
