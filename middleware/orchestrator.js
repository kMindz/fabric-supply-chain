/**
 *
 */

const helper = require('./helper');
const ConfigHelper = helper.ConfigHelper;

module.exports = function (require) {

  const log4js = require('log4js');
  const logger = log4js.getLogger('orchestrator');

  const tools = require('../lib/tools');
  const hfc = require('../lib-fabric/hfc');
  const invoke = require('../lib-fabric/invoke-transaction');
  const peerListener = require('../lib-fabric/peer-listener');

  const configHelper = new ConfigHelper(hfc.getConfigSetting('config'));
  const networkConfig = configHelper.networkConfig;

  const ORG = process.env.ORG || 'a';

  const endorsePeerId = Object.keys(networkConfig[ORG]||{}).filter(k=>k.startsWith('peer'))[0];
  const endorsePeerHost = tools.getHost(networkConfig[ORG][endorsePeerId].requests);

  const USERNAME = process.env.SERVICE_USER || 'service' /*config.user.username*/;


  logger.info('**************    ORCHESTRATOR     ******************');
  logger.info('Admin   \t: ' + USERNAME);
  logger.info('Org name\t: ' + ORG);
  logger.info('Endorse peer\t: (%s) %s', endorsePeerId, endorsePeerHost);
  logger.info('**************                     ******************');


  const TYPE_ENDORSER_TRANSACTION = 'ENDORSER_TRANSACTION';

  ///////////////////////////////////////////
  /// main activity
  logger.info('registering for block events');
  peerListener.registerBlockEvent(function (block) {
    try {
      block.data.data.forEach(blockData => {

        let type = helper.getTransactionType(blockData);
        let channel = helper.getTransactionChannel(blockData);

        logger.info(`got block no. ${block.header.number}: ${type} on channel ${channel}`);

        if (type === TYPE_ENDORSER_TRANSACTION) {

          blockData.payload.data.actions.forEach(action => {
            let extension = action.payload.action.proposal_response_payload.extension;
            let event = extension.events;
            if(!event.event_name) {
              return;
            }
            logger.trace(`event ${event.event_name}`);

            if(event.event_name === 'TransferDetails.Accepted') {
              // instruction is executed, however still has 'matched' status in ledger (but 'executed' in the event)
              let transferDetails = JSON.parse(event.payload.toString());
              logger.trace(event.event_name, JSON.stringify(transferDetails));

              //transferDetails = helper.normalizeInstruction(transferDetails);
              updateProductOwner(transferDetails, transferDetails.new_owner /* 'executed' */);
              return;
            }

            logger.trace('Event not processed:', event.event_name);
          }); // thru action elements
        }
      }); // thru block data elements
    }
    catch(e) {
      logger.error('Caught while processing block event', e);
    }
  });

  /**
   *
   */
  function updateProductOwner(transferDetails, owner) {
    var json = JSON.stringify(transferDetails);
    logger.debug(`set product owner: ${owner} for`, json);

    let channel = 'common';
    logger.debug(`got channel ${channel} for`, json);

    //
    const args = [transferDetails.product_key, transferDetails.old_owner, transferDetails.new_owner, Date.now() + ''];
    return invoke.invokeChaincode([endorsePeerHost], channel, 'reference', 'updateOwner', args, USERNAME, ORG)
      .then(function(/*transactionId*/) {
        logger.info('Update product owner success', transferDetails);
      })
      .catch(function (e) {
        logger.error('Cannot update product owner', transferDetails, e);
      });
  }

};
