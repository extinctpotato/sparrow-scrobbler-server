var apiHelper = {
  currPage: 0,
  apiBaseUrl: 'http://localhost:6789/api',
  tracks: {},
  getTracks: function() {
    var r = new XMLHttpRequest();
    r.open('GET', `${this.apiBaseUrl}/tracks?page=${this.currPage}`, true);
    r.onreadystatechange = function() {
      if(r.readyState == 4) {
        if(r.status == 200) {
          this.tracks = JSON.parse(r.responseText);
        }
      }
    }.bind(this);
    r.send(null);
  },
};
