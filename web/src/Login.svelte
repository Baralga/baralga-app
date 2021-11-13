<script>
  import Api from "./api";
  import { navigate } from "svelte-routing";

  let username;
  let password;

  function login() {
    Api.post("/api/auth/login", { username: username, password: password })
      .then(r => {
        navigate("/", { replace: true });
      })
      .catch(err => {
        return;
      });
  }
</script>

<div class="container">
  <div class="columns is-centered">
    <div class="column is-5-tablet is-4-desktop is-3-widescreen">
      <form class="box" on:submit|preventDefault={login}>
        <div class="field">
          <label for="" class="label">Username</label>
          <div class="control has-icons-left">
            <input
              type="text"
              placeholder="e.g. bobsmith@gmail.com"
              class="input"
              required
              bind:value={username} />
            <span class="icon is-small is-left">
              <i class="fa fa-envelope" />
            </span>
          </div>
        </div>
        <div class="field">
          <label for="" class="label">Password</label>
          <div class="control has-icons-left">
            <input
              type="password"
              placeholder="*******"
              class="input"
              required
              bind:value={password} />
            <span class="icon is-small is-left">
              <i class="fa fa-lock" />
            </span>
          </div>
        </div>
        <!--
            <div class="field">
              <label for="" class="checkbox">
                <input type="checkbox">
               Remember me
              </label>
            </div>
            -->
        <div class="field">
          <button class="button is-success" type="submit">Login</button>
        </div>
      </form>
    </div>
  </div>
</div>
