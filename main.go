package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/mtslzr/pokeapi-go"
)

var (
	Window     *gtk.Window
	RootBox    *gtk.Box
	PokeBox    *gtk.Box
	DexPane    *gtk.Paned
	PokePane   *gtk.Paned
	PokeScroll *gtk.ScrolledWindow
	Entry      *gtk.Entry
	DexView    *gtk.TreeView
	PokeView   *gtk.TreeView
	StatView   *gtk.TreeView
	DexStore   *gtk.ListStore
	PokeStore  *gtk.ListStore
	StatStore  *gtk.ListStore
	PokeImg    *gtk.Image
)

// Appends single value to the DexView's model
func AppendToList(ListStore *gtk.ListStore, column int, value string) {
	ListStore.SetValue(ListStore.Append(), column, value)
}

// Appends several values to the DexView's model
func AppendMultipleToList(ListStore *gtk.ListStore, values ...string) {
	for _, v := range values {
		AppendToList(ListStore, 0, v)
	}
}

func createPixBuf(filename string) *gdk.Pixbuf {
	pixbuf, err := gdk.PixbufNewFromFile(filename)
	if err != nil {
		log.Fatal("Unable to create pixbuf: ", err)
	}
	return pixbuf
}

func downloadURL(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	return body
}

// Handler of "changed" signal of DexView's selection
func DexSelectionChanged(s *gtk.TreeSelection) {
	// Returns glib.List of gtk.TreePath pointers
	rows := s.GetSelectedRows(DexStore)
	items := make([]string, 0, rows.Length())

	for l := rows; l != nil; l = l.Next() {
		path := l.Data().(*gtk.TreePath)
		iter, _ := DexStore.GetIter(path)
		value, _ := DexStore.GetValue(iter, 0)
		str, _ := value.GetString()
		items = append(items, str)
	}

	PokeStore.Clear()
	go func() {
		for _, i := range items {
			g, _ := pokeapi.Pokedex(i)
			for _, p := range g.PokemonEntries {
				PokeStore.Set(PokeStore.Append(),
					[]int{0, 1},
					[]interface{}{p.PokemonSpecies.Name, p.EntryNumber})
			}
		}
	}()

	Entry.SetText(fmt.Sprint(items))
}

// Handler of "changed" signal of PokeView's selection
func PokeSelectionChanged(s *gtk.TreeSelection) {
	// Returns glib.List of gtk.TreePath pointers
	rows := s.GetSelectedRows(PokeStore)

	if rows != nil {
		path := rows.Data().(*gtk.TreePath)
		iter, _ := PokeStore.GetIter(path)
		value, _ := PokeStore.GetValue(iter, 1)
		val, _ := value.GoValue()
		num := fmt.Sprint(val)

		// go func() { // seperate thread to avoid freezes
		p, _ := pokeapi.Pokemon(string(num))

		StatStore.Clear()
		for _, t := range p.Types {
			typeName := t.Type.Name
			i := createPixBuf("imgs/" + typeName + ".png")
			i, _ = i.ScaleSimple(20, 20, gdk.INTERP_BILINEAR)
			StatStore.Set(StatStore.Append(),
				[]int{0, 1, 2},
				[]interface{}{i, "Type", typeName})
		}

		i, _ := gdk.PixbufNewFromDataOnly(downloadURL(p.Sprites.FrontDefault))
		i, _ = i.ScaleSimple(200, 200, gdk.INTERP_TILES)
		PokeImg.SetFromPixbuf(i)

		// }()
	}
}

// Add a column to the tree view (during the initialization of the tree view)
func createTextColumn(title string, id int) *gtk.TreeViewColumn {
	cellRenderer, _ := gtk.CellRendererTextNew()
	column, _ := gtk.TreeViewColumnNewWithAttribute(title, cellRenderer, "text", id)

	return column
}

// Add a column to the tree view (during the initialization of the tree view)
func createPixColumn(title string, id int) *gtk.TreeViewColumn {
	cellRenderer, _ := gtk.CellRendererPixbufNew()
	column, _ := gtk.TreeViewColumnNewWithAttribute(title, cellRenderer, "pixbuf", id)

	return column
}

func main() {
	gtk.Init(&os.Args)

	// Declarations
	Window, _ = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	RootBox, _ = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	PokeBox, _ = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	DexPane, _ = gtk.PanedNew(gtk.ORIENTATION_HORIZONTAL)
	PokePane, _ = gtk.PanedNew(gtk.ORIENTATION_HORIZONTAL)
	PokeScroll, _ = gtk.ScrolledWindowNew(nil, nil)
	DexView, _ = gtk.TreeViewNew()
	PokeView, _ = gtk.TreeViewNew()
	StatView, _ = gtk.TreeViewNew()
	Entry, _ = gtk.EntryNew()
	DexStore, _ = gtk.ListStoreNew(glib.TYPE_STRING)
	PokeStore, _ = gtk.ListStoreNew(glib.TYPE_STRING, glib.TYPE_INT)
	StatStore, _ = gtk.ListStoreNew(glib.TYPE_OBJECT, glib.TYPE_STRING, glib.TYPE_STRING)
	PokeImg, _ = gtk.ImageNew()

	// Window properties
	Window.SetTitle("Pokedex GTK")
	Window.Connect("destroy", gtk.MainQuit)
	Window.SetDefaultSize(1000, 750)

	p, _ := gdk.PixbufNewFromFile("imgs/icon.png")
	Window.SetIcon(p)

	// DexView properties
	{
		DexView.AppendColumn(createTextColumn("Pokedex", 0))
	}
	DexView.SetModel(DexStore)
	// PokeView properties
	{
		PokeView.AppendColumn(createTextColumn("#", 1))
		PokeView.AppendColumn(createTextColumn("Pokemon Name", 0))
	}
	PokeView.SetModel(PokeStore)

	{
		StatView.AppendColumn(createPixColumn("Icon", 0))
		StatView.AppendColumn(createTextColumn("Stat Name", 1))
		StatView.AppendColumn(createTextColumn("Stat", 2))
	}
	StatView.SetModel(StatStore)

	// DexView selection properties
	dexsel, _ := DexView.GetSelection()
	dexsel.SetMode(gtk.SELECTION_SINGLE)
	dexsel.Connect("changed", DexSelectionChanged)

	// PokeView selection properties
	pokesel, _ := PokeView.GetSelection()
	pokesel.SetMode(gtk.SELECTION_SINGLE)
	pokesel.Connect("changed", PokeSelectionChanged)

	// DexPane properties
	DexPane.SetWideHandle(true)
	DexPane.SetPosition(0)

	// PokePane properties
	PokePane.SetPosition(150)

	Entry.SetEditable(false)

	// Packing
	PokeScroll.Add(PokeView)
	PokeBox.PackStart(PokeImg, true, true, 0)
	PokeBox.PackStart(StatView, true, true, 0)
	PokePane.Add1(PokeScroll)
	PokePane.Add2(PokeBox)
	DexPane.Add1(DexView)
	DexPane.Add2(PokePane)
	RootBox.PackStart(DexPane, true, true, 0)
	RootBox.PackStart(Entry, false, false, 0)
	Window.Add(RootBox)

	g, _ := pokeapi.Resource("pokedex")
	for _, i := range g.Results {
		AppendToList(DexStore, 0, i.Name)
	}

	i, _ := DexStore.GetIterFirst()
	dexsel.SelectIter(i)

	Window.ShowAll()
	gtk.Main()
}
