var winners = []


function teamClick(elem){
  //console.log("element", elem)
  let team = elem.innerText
  let parent = elem.parentElement
  let pos = elem.getAttribute('pos')
  currSelected = parent.getAttribute("winner")

  if (currSelected == pos) {
    parent.setAttribute("winner", "")
    winners = winners.filter((word) => word != team)
  } else {
    parent.setAttribute("winner", pos)
    winners.push(team)
  }

  let children = elem.parentElement.children
  console.log(children)
}
