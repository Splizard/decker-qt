import QtQuick 2.2
import QtQuick.Controls 1.1
import QtQuick.Dialogs 1.0
import QtQuick.Layouts 1.0
import "qml"
  
ApplicationWindow {
    visible: true
    title: "Decker"

    width: 640
    height: 420
    minimumHeight: 400
    minimumWidth: 600 
    
    
    FileDialog {
        id: fileDialog
        nameFilters: [ "Deck files (*.deck *.txt)" ]
        onAccepted: 
        	cards.open(fileUrl),
        	dummyModel.update()
    }
   
   Action {
        id: openAction
        text: "&Open"
        shortcut: StandardKey.Open
        iconName: "document-open"
        onTriggered: fileDialog.open()
        tooltip: "Open"
    }
    
     Action {
        id: saveAction
        text: "&Save"
        shortcut: StandardKey.Save
        iconName: "document-save"
        onTriggered: cards.save()
        tooltip: "Save"
    }
    
    Action {
        id: addAction
        text: "&Edit"
        iconName: "add"
        onTriggered: add()
        function add() {
    		dummyModel.append({
				"num": "1", 
				"title": "" 
			})
			cards.add()
        }
        tooltip: "Add card"
    }
    
     Action {
        id: removeAction
        text: "&Edit"
        iconName: "gtk-delete"
        onTriggered: remove(tableview.currentRow)
        function remove(i) {
    		dummyModel.remove(i)
			cards.remove(i)
        }
        tooltip: "Add card"
    }
    
    toolBar: ToolBar {
        id: toolbar
        RowLayout {
            id: toolbarLayout
            spacing: 4
            width: parent.width
            ToolButton { action: openAction }
            ToolButton { action: saveAction }
            ToolButton { action: addAction }
            ToolButton { action: removeAction }
            Item { Layout.fillWidth: true }
        }
    }
		
	ListModel {
		id: dummyModel
		Component.onCompleted: {
			update()
		}
		
		function update()
		{
			dummyModel.clear()
			for (var i = 0 ; i < cards.len() ; ++i) {
				dummyModel.append({
					"num": cards.amount(i), 
					"title": cards.name(i) 
				})
			}
			game.text = "Game: "+cards.game()+"  Total: "+cards.total()
			cardImage.source = ""
			tableview.selection.clear()
			tableview.currentRow = 0
		}
	}
	
 	Text {
 		id: game
		text: "Game: "+cards.game()+"  Total: "+cards.total()
		anchors.left:  parent.left
		anchors.top: parent.top
		anchors.margins: 8
		//cards.name(tableview.currentRow)
	}

	SplitView {
	    anchors.top: game.bottom
	    anchors.bottom: parent.bottom
		anchors.left:  parent.left
		anchors.right:  parent.right
	    anchors.margins: 8
	    Layout.fillWidth: true
    
		TableView{
			id: tableview
			model: dummyModel
			
			width: 300
			
			anchors.bottom: parent.bottom
			anchors.left:  parent.left

			TableViewColumn {
				role: "num"
				title: "#of"
				width: 36
				resizable: false
				movable: false
		
			}
			TableViewColumn {
				role: "title"
				title: "Card Name"
				width: 240
			}
			
			selection.onSelectionChanged: {
				cards.load(tableview.currentRow),
				imageTimer.start()
				loading.start()
				if (tableview.currentRow > -1) {
					loading.visible =  true
				}
			}
			
			property bool editing
			
			Component {
				id: textComponent
				Text{
					text:styleData.value
				}
			}
		 
		 
			Component {
				id: editComponent
				TextInput{
					id: textinput
					text: styleData.value
					selectByMouse: true
					onAccepted: edit_name()
					
					activeFocusOnPress: false
					
					color: text == styleData.value ? "black" : Qt.rgba(0.3, 0.3, 0.3, 1)
					
					function edit_name() {
						if (styleData.column == 0) {
							model.setProperty(styleData.row, "num", text)
							cards.setamount(styleData.row, text)
							game.text = "Game: "+cards.game()+"  Total: "+cards.total()
						} else {
							model.setProperty(styleData.row, "title", text)
							cards.setname(styleData.row, text)
						}
						cards.load(styleData.row),
						imageTimer.start()
						loading.start()
						loading.visible =  true
					}
					
					function on_click() {
						tableview.selection.clear()
						tableview.selection.select(styleData.row)
						textinput.forceActiveFocus()
						cards.load(tableview.currentRow),
						imageTimer.start()
						loading.start()
						loading.visible = true
					}
					
					MouseArea {
                        id: mouseArea
                        anchors.fill: parent
                         focus: false
                        hoverEnabled: false
                        onClicked: on_click()
                        preventStealing: true
                    }
				}
			}
 

			
			itemDelegate: editComponent

			
			function edit() {
				editing = !editing
				if (editing) {
				
					itemDelegate = editComponent
					
				} else {
					itemDelegate = textComponent
				}
			}
			
		}
		
		Timer {
			id: imageTimer
			interval: 50; running: true; repeat: true;
			
			function step() {
				game.text = "Game: "+cards.game()+"  Total: "+cards.total()
				if (cards.loaded(tableview.currentRow) == true) {
					loading.stop()
					loading.visible = false,
					cardImage.source = "image://card/"+cards.image(tableview.currentRow)
					imageTimer.stop()
				}
			}
			
			onTriggered: step()
				
				
		}
		SplitView {
	        orientation: Qt.Vertical
			Image {
				width: 150
				id: cardImage
				source: ""
				asynchronous: true
			
				anchors.top: parent.top
				anchors.bottom: parent.bottom
				anchors.right: parent.right
			
				fillMode: Image.PreserveAspectFit
			
				Rectangle {
					z: 1
					anchors.centerIn: parent
					LoadCircle {
						id: loading
						visible: false
						anchors.centerIn: parent
					}
				}
			}
		}
	}
}
