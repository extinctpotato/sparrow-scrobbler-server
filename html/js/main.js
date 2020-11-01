var apiHelper = {
  currPage: 0,
  maxPage: 0,
  apiBaseUrl: `${window.location.href}/api`,
  transform: {
    tag: 'tr',
    class: 'track-tr',
    children: [
      {
        'tag':'td',
        'html':'${id}',
      },
      {
        'tag':'td',
        'html':'${artist}',
      },
      {
        'tag':'td',
        'html':'${album}',
      },
      {
        'tag':'td',
        'html':'${name}',
      },
      {
        'tag':'td',
        'html':'${played_at}',
      },
    ],
  },
  tracks: {},
  getTracks: function() {
    var r = new XMLHttpRequest();
    r.open('GET', `${this.apiBaseUrl}/tracks?page=${this.currPage}`, true);
    r.onreadystatechange = function() {
      if(r.readyState == 4) {
        if(r.status == 200) {
          this.tracks = JSON.parse(r.responseText);

          if(this.maxPage === 0) {
            this.maxPage = Math.floor(apiHelper.tracks[0].id/30);
          }

          if(this.currPage === 0) {
            $('#prev-page-btn').addClass('disabled');
          } else {
            $('#prev-page-btn').removeClass('disabled');
          }

          if (this.currPage === this.maxPage) {
            $('#next-page-btn').addClass('disabled');
          } else {
            $('#next-page-btn').removeClass('disabled');
          }

          $('.track-tr').remove();
          $('#tracks').json2html(this.tracks, this.transform);
        }
      }
    }.bind(this);
    r.send(null);
  },
  nextPage: function() {
    this.currPage++;
    this.getTracks();
  },
  prevPage: function() {
    this.currPage--;
    this.getTracks();
  },
};

window.addEventListener('DOMContentLoaded', (event) => {
  apiHelper.getTracks();

  $('ul.pagination li a').on('click', function(e) {
    e.preventDefault();
    var tag = $(this).text();

    if(tag === 'Next') {
      apiHelper.nextPage();
    } else if(tag === 'Previous') {
      apiHelper.prevPage();
    }

    $("html, body").animate({ scrollTop: 0 }, "slow");
  });
});
