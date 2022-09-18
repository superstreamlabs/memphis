// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import './style.scss';

import { makeStyles, withStyles } from '@material-ui/core/styles';
import DialogContent from '@material-ui/core/DialogContent';
import Dialog from '@material-ui/core/Dialog';
import Tabs from '@material-ui/core/Tabs';
import Tab from '@material-ui/core/Tab';
import React, { useState } from 'react';

import FunctionsOverview from '../../components/functionsOverview';
import { getFontColor } from '../../utils/styleTemplates';
import Seperator from '../../assets/images/seperator.svg';
import ConnectToHub from '../../components/connectToHub';
import Connect from '../../assets/images/connect.svg';
import Button from '../../components/button';
import Close from '../../assets/images/close.svg';
import FunctionDetails from './functionDetails';
import FunctionsList from './functionsList';

const AntTabs = withStyles({
    root: {
        paddingLeft: '30px'
    },
    indicator: {
        backgroundColor: getFontColor('black')
    }
})(Tabs);

const functions = [
    {
        _id: 1,
        name: 'sveta sveta sveta sveta sveta sveta sveta sveta',
        type: 'blabl'
    },
    {
        _id: 2,
        name: 'sveta2',
        type: 'blabl'
    },
    {
        _id: 3,
        name: 'sveta3',
        type: 'blabl'
    }
];

const AntTab = withStyles((theme) => ({
    root: {
        textTransform: 'none',
        fontSize: '16px',
        minWidth: 12,
        maxWidth: 100,
        fontWeight: theme.typography.fontWeightBold,
        marginRight: theme.spacing(3),
        fontFamily: ['Inter'].join(','),
        '&:hover': {
            color: getFontColor('navy'),
            opacity: 1
        },
        '&$selected': {
            color: getFontColor('navy'),
            fontWeight: theme.typography.fontWeightBold
        },
        '&:focus': {
            color: getFontColor('navy')
        }
    },
    selected: {}
}))((props) => <Tab disableRipple {...props} />);

const useStyles = makeStyles((theme) => ({
    root: {
        flexGrow: 1
    },
    dialogPaper: {
        height: '100vh',
        width: '100vw',
        borderRadius: '10px',
        minWidth: '1400px',
        overflowX: 'hidden',
        overflowY: 'scroll',
        position: 'relative'
    },
    dialogContent: {
        width: '100%',
        height: '100%',
        padding: '0px'
    }
}));

function HubMarketplace(props) {
    const classes = useStyles();
    const [value, setValue] = useState(0);
    const [chosenFunction, setChosenFunction] = useState(null);
    const [signedToHub, signInToHub] = useState(false); //placeholder - will be received from state
    const [isOpenSignIn, flipIsOpenSignIn] = useState(false);

    const handleChangeMenuItem = (_, newValue) => {
        setValue(newValue);
    };

    const handleCloseModal = () => {
        props.closeModal(false);
    };

    return (
        <Dialog
            open={props.open}
            onClose={(_, reson) => {
                if (reson === 'backdropClick') handleCloseModal();
                // { props.clickOutside() }
            }}
            classes={{ paper: classes.dialogPaper }}
        >
            <ConnectToHub open={isOpenSignIn} closeModeal={(e) => flipIsOpenSignIn(e)} />
            <div className="functions-modal-header">
                <div>
                    <label className="visit-hub">Visit hub</label>
                    <img src={Seperator} alt="|" width="20" height="20" className="seperator" />
                    {signedToHub ? (
                        <>
                            <img src={Connect} alt="connect to hub" width="20" height="20" className="pointer" />
                            <label className="sign-in-hub connected">Connected to hub</label>
                        </>
                    ) : (
                        <label className="sign-in-hub" onClick={() => flipIsOpenSignIn(true)}>
                            Sign in to hub
                        </label>
                    )}

                    <img src={Close} alt="close" width="12" height="12" style={{ cursor: 'pointer' }} onClick={handleCloseModal} />
                </div>
            </div>

            <DialogContent className={classes.dialogContent}>
                <AntTabs value={value} onChange={handleChangeMenuItem}>
                    <AntTab label="Private" />
                    <AntTab label="Public" />
                </AntTabs>
                <div className="functions-container">
                    <div className="function-list-section">
                        <FunctionsList chooseFunction={(func) => setChosenFunction(func)} />
                    </div>
                    <div className="function-details-section">
                        {' '}
                        <FunctionDetails chosenFunction={chosenFunction} />
                    </div>
                    {/* {value === 0 && <h1>Private</h1>}
                    {value === 1 && <h1>Public</h1>} */}
                </div>
                <div className="chosen-func-container">
                    <FunctionsOverview functions={functions} horizontal={true} editable={true} />
                </div>
                <div className="func-btn-footer">
                    <Button
                        className="modal-btn"
                        width="90px"
                        height="32px"
                        placeholder="Cancel"
                        colorType="purple"
                        radiusType="circle"
                        backgroundColorType="none"
                        fontSize="14px"
                        fontWeight="bold"
                        aria-haspopup="true"
                        // onClick={() => setOpenFunctionForm(true)}
                    />
                    <Button
                        className="modal-btn"
                        width="90px"
                        height="32px"
                        placeholder="Add"
                        colorType="white"
                        radiusType="circle"
                        backgroundColorType="purple"
                        fontSize="14px"
                        fontWeight="bold"
                        aria-haspopup="true"
                        // onClick={() => setOpenFunctionForm(true)}
                    />
                </div>
            </DialogContent>
        </Dialog>
    );
}
export default HubMarketplace;
