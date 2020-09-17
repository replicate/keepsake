import { Tabs, TabList, Tab, TabPanels, TabPanel } from "@reach/tabs";

const TabContext = React.createContext({ setState: () => {} });

function StatefulTabs({ children, ...props }) {
  return (
    <TabContext.Consumer>
      {({ setTabIndex, activeTabIndex }) => (
        <Tabs
          className="tabs"
          index={activeTabIndex}
          onChange={setTabIndex}
          {...props}
        >
          {children}
        </Tabs>
      )}
    </TabContext.Consumer>
  );
}

class TabState extends React.Component {
  state = {
    activeTabIndex: 0,
    setTabIndex: (tabIndex) => this.setState({ activeTabIndex: tabIndex }),
  };

  render() {
    return (
      <TabContext.Provider value={this.state}>
        {this.props.children}
      </TabContext.Provider>
    );
  }
}

export { StatefulTabs, TabList, Tab, TabPanels, TabPanel, TabState };
