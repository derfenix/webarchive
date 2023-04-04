function index() {
  $.ajax({
    url: "/api/v1/pages", success: function (data, status, xhr) {
      if (status !== "success") {
        gotError(status);

        return;
      }

      let elem = document.getElementById("data");
      elem.innerHTML = "";
      // elem.attachShadow({mode: 'open'});

      data.forEach(function (v) {
        let page_elem = pages_tmpl.content.cloneNode(true);
        $(page_elem).find(".url").attr("onclick", "goToPage('" + v.id + "');");
        $(page_elem).find(".status").addClass(v.status);
        $(page_elem).find(".status").attr("title", v.status);
        $(page_elem).find(".created").html(v.created);
        $(page_elem).find(".title").html(v.meta.title);
        $(page_elem).find(".description").html(v.meta.description);
        elem.append(page_elem); // (*)
      })
    }
  })
}

function goToPage(id) {
  history.pushState({"page": id}, null, id);
  page(id);
}

function page(id) {
  $.ajax({
    url: "/api/v1/pages/" + id, success: function (data, status, xhr) {
      if (status !== "success") {
        gotError(status);

        return;
      }

      let elem = document.getElementById("data");
      elem.innerHTML = "";
      let page_elem = page_tmpl.content.cloneNode(true);
      $(page_elem).find("#page_title").html(data.meta.title);
      $(page_elem).find("#page_description").html(data.meta.description);
      $(page_elem).find("#page_url").html(data.url);

      data.results.forEach(function (result) {
        let result_elem = result_tmpl.content.cloneNode(true);
        $(result_elem).find(".format").html(result.format);
        if (result.error !== "" && result.error !== undefined) {
          $(result_elem).find(".format").addClass("error");
          $(result_elem).find(".result_link").html("⚠");
          $(result_elem).find(".result_link").attr("title", result.error);
        } else {
          result.files.forEach(function (file) {
            $(result_elem).find(".result_link").attr("onclick", "window.open('/api/v1/pages/" + data.id + "/file/" + file.id + "', '_blank');");
            $(result_elem).find(".result_link").html(file.name);
          })
        }

        $(page_elem).find("#results").append(result_elem);
      })

      elem.append(page_elem); // (*)
    }
  })
}

function gotError(err) {
  console.log(err);
}

document.addEventListener("DOMContentLoaded", function () {
  $("#site_title").html("WebArchive " + window.location.hostname);
  document.title = "WebArchive " + window.location.hostname;
  if (window.location.pathname.endsWith("/")) {
    index();
  } else {
    page(window.location.pathname.slice(1));
  }
});
window.addEventListener('popstate', function (event) {
  if (event.state === null) {
    index();
  } else {
    page(event.state.page);
  }
});
