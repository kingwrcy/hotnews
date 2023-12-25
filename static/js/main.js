$(function () {
    $("input[name='result']").click(function () {
        const val = $(this).val()
        const index = $(this).data('index')
        if (val === 'reject') {
            $(`#reason-${index}`).show()
        } else {
            $(`#reason-${index}`).hide()
        }
    })

    const $sidebar = $("#sidebar");
    $("#showSidebar").click((e)=>{
        $sidebar.show().fadeIn()
        e.stopPropagation()
    })

    $(document.body).click(()=>{
        console.log($sidebar.css('display'))
      if ($sidebar.css('display') === 'block'){
          $sidebar.hide().fadeOut()
      }
    })
})