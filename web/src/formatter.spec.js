import {
  formatDuration,
  asFormattedDuration,
  asFilterLabel,
} from "./formatter.js";
import moment from "moment/src/moment";
import assert from "assert";

describe("Formatter Tests", () => {
  describe("Format Duration", () => {
    it("should format the duration with hours", () => {
      let duration = moment.duration(2, "hours");
      assert.equal(formatDuration(duration), "02:00");
    });

    it("should format the duration with minutes", () => {
      let duration = moment.duration(30, "minutes");
      assert.equal(formatDuration(duration), "00:30");
    });
  });

  describe("Format Filter Label", () => {
    it("should format the filter label for year", () => {
      let filter = {
        timespan: "year",
        from: moment("2020-11-01"),
      };
      assert.equal(asFilterLabel(filter), "2020");
    });

    it("should format the filter label for quarter", () => {
      let filter = {
        timespan: "quarter",
        from: moment("2020-11-01"),
      };
      assert.equal(asFilterLabel(filter), "4th Quarter 2020");
    });

    it("should format the filter label for month", () => {
      let filter = {
        timespan: "month",
        from: moment("2020-11-01"),
      };
      assert.equal(asFilterLabel(filter), "November 2020");
    });

    it("should format the filter label for week", () => {
      let filter = {
        timespan: "week",
        from: moment("2020-11-01"),
      };
      assert.equal(asFilterLabel(filter), "45th Week November 2020");
    });
  });
});
