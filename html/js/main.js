var apiHelper = {
  currPage: 0,
  apiBaseUrl: 'http://localhost:6789/api',
  transform: {
    tag: 'tr',
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
          $('#tracks tbody').remove();
          $('#tracks').json2html(this.tracks, this.transform);
        }
      }
    }.bind(this);
    r.send(null);
  },
};

window.addEventListener('DOMContentLoaded', (event) => {
    apiHelper.getTracks();
});
