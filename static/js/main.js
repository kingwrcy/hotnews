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
})