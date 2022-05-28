document.addEventListener('DOMContentLoaded', function() {
    document.body.addEventListener('baralga__main_content_modal-show', function (evt) {
        var modal = bootstrap.Modal.getOrCreateInstance(document.getElementById('baralga__main_content_modal'), { keyboard: true });
        modal.show();
    });
    document.body.addEventListener('baralga__main_content_modal-hide', function (evt) {
        var modal = bootstrap.Modal.getOrCreateInstance(document.getElementById('baralga__main_content_modal'), { keyboard: true });
        modal.hide();
    });
});